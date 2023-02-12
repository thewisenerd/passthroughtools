package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var (
	debug = false

	port = 8002

	hostPublic = "passthroughtools.org"
)

func cpuPin(w http.ResponseWriter, r *http.Request) {
	if !debug {
		if r.Host != hostPublic {
			log.Printf("got invalid host, host=%s\n", r.Host)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	if r.Method != http.MethodPost {
		log.Printf("got invalid method, method=%s\n", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vcpuStr := strings.Clone(r.FormValue("vcpu"))
	lscpuStr := strings.Clone(r.FormValue("lscpu"))

	vcpu, err := strconv.Atoi(vcpuStr)
	if err != nil {
		log.Printf("unable to parse vcpu, value=%s\n", vcpuStr)
		http.Error(w, "Bad Request, unable to parse vcpu value", http.StatusBadRequest)
		return
	}

	lscpuEncoded := base64.StdEncoding.EncodeToString([]byte(lscpuStr))
	log.Printf("received request, vcpu=%d, lscpu=%s", vcpu, lscpuEncoded)

	suggestion, err := Suggest(lscpuStr, vcpu)
	if err != nil {
		log.Printf("unable to suggest response, err=%v\n", err)
		http.Error(w, fmt.Sprintf("Internal Server Error, unable to suggest response, error=%s", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("suggestion=%v", suggestion)

	wb, err := FormatSuggestion(suggestion)
	if err != nil {
		log.Printf("unable to format suggestion, err=%v\n", err)
		http.Error(w, fmt.Sprintf("Internal Server Error, unable to format suggestion, error=%s", err), http.StatusInternalServerError)
	}

	_, err = io.WriteString(w, *wb)
	if err != nil {
		log.Printf("unable to write string, err=%v\n", err)
	}
}

func main() {
	flag.BoolVar(&debug, "d", false, "debug mode")
	flag.Parse()

	log.Printf("starting server at port=%d, debug=%v\n", port, debug)

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/cpupin/", cpuPin)
	server := &http.Server{Addr: "127.0.0.1:" + strconv.Itoa(port), Handler: mux}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("http.ListenAndServe failed", err)
	}
}
