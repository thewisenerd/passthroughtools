package main

import (
	"errors"
	"fmt"
	"log"
	"math/bits"
	"sort"
	"strconv"
	"strings"
)

type CpuLine struct {
	cpu    int
	core   int
	socket int
	node   int
}

func parseIntOrNil(s string) *int {
	d, err := strconv.Atoi(s)
	if err != nil {
		return nil
	}
	return &d
}

func vIntOrNil(d *int) string {
	if d == nil {
		return "(nil)"
	}
	return strconv.Itoa(*d)
}

func parseLine(parts []string) (*CpuLine, error) {
	if len(parts) <= 4 {
		return nil, errors.New("insufficient parts")
	}

	cpu := parseIntOrNil(parts[0])
	core := parseIntOrNil(parts[1])
	socket := parseIntOrNil(parts[2])
	node := parseIntOrNil(parts[3])

	if cpu == nil || core == nil || socket == nil || node == nil {
		return nil, errors.New(fmt.Sprintf("failed to parse parts, cpu=%s, core=%s, socket=%s, node=%s", vIntOrNil(cpu), vIntOrNil(core), vIntOrNil(socket), vIntOrNil(node)))
	}

	return &CpuLine{
		*cpu,
		*core,
		*socket,
		*node,
	}, nil
}

func parse(lscpu string) (*[]CpuLine, error) {
	lines := strings.Split(lscpu, "\n")
	var cpuLines []CpuLine
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		line = strings.TrimSpace(line)

		if len(line) == 0 || line[0] == '#' {
			continue
		}

		log.Printf("got line=%s\n", line)
		parts := strings.Split(line, ",")
		cpuLine, err := parseLine(parts)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse linenumber=%d, line=%s, error=%v", i+1, line, err))
		}

		log.Printf("parsed line=%v, weight=%d", *cpuLine, lineWeight(*cpuLine))
		cpuLines = append(cpuLines, *cpuLine)
	}

	return &cpuLines, nil
}

// plus one's to avoid multiplication by zero
func lineWeight(line CpuLine) uint64 {
	var ret uint64 = 0
	var rot = 8 // "arbitrary"

	ret += uint64(line.node + 1)
	ret = bits.RotateLeft64(ret, rot)

	ret += uint64(line.socket + 1)
	ret = bits.RotateLeft64(ret, rot)

	ret += uint64(line.core + 1)
	ret = bits.RotateLeft64(ret, rot)

	ret += uint64(line.cpu + 1)

	return ret
}

// Suggest TODO: use structured inputs
func Suggest(lscpu string, vcpu int) (*[]CpuLine, error) {
	topology, err := parse(lscpu)
	topologyLen := len(*topology)
	if err != nil {
		return nil, err
	}
	fmt.Printf("got topology=%v\n", topology)

	if vcpu > topologyLen {
		return nil, errors.New(fmt.Sprintf("cannot request more vcpu than available cpus, vcpu=%d, available=%d", vcpu, len(*topology)))
	}

	// simplest logic for now
	sort.Slice(*topology, func(i, j int) bool {
		return lineWeight((*topology)[i]) < lineWeight((*topology)[j])
	})
	fmt.Printf("sorted topology=%v\n", topology)

	suggestion := (*topology)[topologyLen-vcpu : topologyLen]
	return &suggestion, nil
}

func FormatSuggestion(suggestion *[]CpuLine) (*string, error) {
	if len(*suggestion) == 0 {
		return nil, errors.New("cannot format empty suggestion")
	}

	sb := ""
	padding := "  "

	sb += fmt.Sprintf("%s<vcpu placement='static'>%d</vcpu>\n", padding, len(*suggestion))
	sb += fmt.Sprintf("%s<cputune>\n", padding)

	// TODO: iffy.. should add a couple validations here.
	//       when core values repeat over node+socket, this will break.
	var socketMap = make(map[int]int)
	var coreMap = make(map[int]int)

	getOrDefault := func(m map[int]int, k int, def int) int {
		if v, ok := m[k]; ok {
			return v
		}
		return def
	}

	for i := 0; i < len(*suggestion); i++ {
		line := (*suggestion)[i]

		sb += fmt.Sprintf("%s%s<vcpupin vcpu='%d' cpuset='%d'/>\n", padding, padding, i, line.cpu)
		socketMap[line.socket] = getOrDefault(socketMap, line.socket, 0) + 1
		coreMap[line.core] = getOrDefault(coreMap, line.core, 0) + 1
	}

	maxThreads := 0
	for _, threadCount := range coreMap {
		if threadCount > maxThreads {
			maxThreads = threadCount
		}
	}

	sb += fmt.Sprintf("%s</cputune>\n", padding)
	sb += fmt.Sprintf("%s<cpu mode='host-passthrough' check='none'>\n", padding)
	sb += fmt.Sprintf("%s%s<topology sockets='%d' cores='%d' threads='%d'/>\n", padding, padding, len(socketMap), len(coreMap), maxThreads)
	sb += fmt.Sprintf("%s%s<cache mode='passthrough'/>\n", padding, padding)
	sb += fmt.Sprintf("%s</cpu>\n", padding)

	return &sb, nil
}
