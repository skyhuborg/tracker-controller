package main

import (
	"fmt"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

func main() {
	v, _ := mem.VirtualMemory()
	h, _ := host.Info()
	c, _ := cpu.Info()
	l, _ := load.Avg()
	d, _ := disk.Partitions(false)
	n, _ := net.Interfaces()

	fmt.Printf("%s\n\n", h)
	fmt.Printf("%s\n\n", v)
	fmt.Printf("%s\n\n", c)
	fmt.Printf("%s\n\n", l)
	fmt.Printf("%s\n\n", d)
	fmt.Printf("%s\n\n", n)

}
