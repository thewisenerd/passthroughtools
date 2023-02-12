#### Tool 1: CPU Pinning Helper (quick, autogenerates config for you)
The CPU Pinning Helper will choose the right cores for you pretty quick. They have a web tool you can use from the browser and an API you can use from terminal.

##### API Method
To use the API simply run the following command, substituting *$CORES* with the number of cores you've assigned your vm:
```sh
$ curl -X POST -F "vcpu=$CORES" -F "lscpu=`lscpu -p`" https://passthroughtools.org/api/v1/cpupin/
```
I'm assigning mine 8 cores, so this is my command:
```sh
$ curl -X POST -F "vcpu=8" -F "lscpu=`lscpu -p`" https://passthroughtools.org/api/v1/cpupin/
```
##### Web Method
Open their web tool by visiting [this link](https://passthroughtools.org/cpupin/). In the first field enter the number of cores you'll be assigning the VM. In the text box below it paste the output of running "lscpu -p" on your host machine. Click **Submit** to have an optimal pinning configuration generated for you.
[![2020-09-12-20-25.png](2020-09-12-20-25.png)](2020-09-12-20-25.png)

Both methods should produce the same results. This is the configuration generated for my setup:
```
  <vcpu placement='static'>8</vcpu>
  <cputune>
    <vcpupin vcpu='0' cpuset='2'/>
    <vcpupin vcpu='1' cpuset='8'/>
    <vcpupin vcpu='2' cpuset='3'/>
    <vcpupin vcpu='3' cpuset='9'/>
    <vcpupin vcpu='4' cpuset='4'/>
    <vcpupin vcpu='5' cpuset='10'/>
    <vcpupin vcpu='6' cpuset='5'/>
    <vcpupin vcpu='7' cpuset='11'/>
  </cputune>
  <cpu mode='host-passthrough' check='none'>
    <topology sockets='1' cores='4' threads='2'/>
    <cache mode='passthrough'/>
  </cpu>
```
Copy the configuration and add it to your VM's XML file. That is it.I recommend you still go through the second tool for a more in-depth understanding of what is going on here. However, you don't have to :-)

