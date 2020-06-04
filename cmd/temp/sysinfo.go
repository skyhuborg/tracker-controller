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

	fmt.Println(h)
	fmt.Println(v)
	fmt.Println(c)
	fmt.Println(l)
	fmt.Printf("%s\n\n", d)
	fmt.Printf("%s\n\n", n)

}
