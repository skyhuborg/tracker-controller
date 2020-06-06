/*
MIT License
-----------

Copyright (c) 2020 Steve McDaniel, Corey Gaspard

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/grandcat/zeroconf"
)

func main() {
	// Discover all services on the network (e.g. _workstation._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			var ports []string
			ports = append(ports, "555")
			ports = append(ports, "554")
			tcpGather(entry.HostName, entry.AddrIPv4[0].String(), ports)
			//log.Println(entry)
		}
		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	err = resolver.Browse(ctx, "_http._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
}

func tcpGather(name string, ip string, ports []string) map[string]string {
	// check emqx 1883, 8083 port

	results := make(map[string]string)
	for _, port := range ports {
		address := net.JoinHostPort(ip, port)
		// 3 second timeout
		conn, err := net.DialTimeout("tcp", address, 3*time.Second)
		if err != nil {
			results[port] = "failed"

			// todo log handler
		} else {
			if conn != nil {
				log.Printf("IP Camera %s found at Address:%s", name, address)
				results[port] = "success"
				_ = conn.Close()
			} else {
				results[port] = "failed"
			}
		}
	}
	return results
}

// package main

// import (
// 	"log"
// 	"os"
// 	"os/signal"
// 	"time"

// 	"github.com/oleksandr/bonjour"
// )

// func main() {
// 	resolver, err := bonjour.NewResolver(nil)
// 	if err != nil {
// 		log.Println("Failed to initialize resolver:", err.Error())
// 		os.Exit(1)
// 	}

// 	results := make(chan *bonjour.ServiceEntry)

// 	go func(results chan *bonjour.ServiceEntry, exitCh chan<- bool) {
// 		for e := range results {
// 			log.Printf("%s, %s - %s(%s):%d", e.Instance, e.ServiceRecord, e.AddrIPv4, e.AddrIPv6, e.Port)
// 			// exitCh <- true
// 			// time.Sleep(1e9)
// 			// os.Exit(0)
// 		}
// 	}(results, resolver.Exit)

// 	err = resolver.Browse("_http._tcp", "local.", results)
// 	if err != nil {
// 		log.Println("Failed to browse:", err.Error())
// 	}

// 	select {}
// }

// func register() {
// 	s, err := bonjour.Register("Sky Hub", "_http._tcp", "", 9999, []string{"txtv=1", "app=test"}, nil)
// 	if err != nil {
// 		log.Fatalln(err.Error())
// 	}

// 	// Ctrl+C handling
// 	handler := make(chan os.Signal, 1)
// 	signal.Notify(handler, os.Interrupt)
// 	for sig := range handler {
// 		if sig == os.Interrupt {
// 			s.Shutdown()
// 			time.Sleep(1e9)
// 			break
// 		}
// 	}
// }

// package main

// import (
// 	"github.com/hashicorp/mdns"
// )

// // type Server struct {
// // 	Handle         *grpc.Server
// // 	ListenPort     int
// // 	EnableTls      bool
// // 	TlsKey         string
// // 	TlsCert        string
// // 	ConfigFile     string
// // 	DbPath         string
// // 	StaticDataPath string
// // 	PipeFilePath   string
// // 	StaticDataPort int
// // 	AuthTokens     []Auth

// // 	config common.Config
// // 	db     trackerdb.DB
// // }

// func main() {
// 	// Setup our service export
// 	host, _ := os.Hostname()
// 	info := []string{"Sky Hub Tracker"}
// 	service, _ := mdns.NewMDNSService(host, "_grpc._tcp", "", "", 8000, nil, info)

// 	// Create the mDNS server, defer shutdown
// 	server, _ := mdns.NewServer(&mdns.Config{Zone: service})
// 	defer server.Shutdown()

// 	// Make a channel for results and start listening
// 	entriesCh := make(chan *mdns.ServiceEntry, 4)
// 	go func() {
// 		for entry := range entriesCh {
// 			fmt.Printf("Got new entry: %v\n", entry)
// 		}
// 	}()

// 	// Start the lookup
// 	mdns.Lookup("_foobar._tcp", entriesCh)
// 	close(entriesCh)

// }
