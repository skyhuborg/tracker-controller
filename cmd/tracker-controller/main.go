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
	"flag"
	"fmt"
	"log"

	"gitlab.com/skyhuborg/tracker-controller/internal/controller"

	//"net/http"
	"os"
	"strconv"
	"time"
)

type Environment struct {
	GrpcListenPort int
	GrpcEnableTls  bool
	GrpcTlsKey     string
	GrpcTlsCert    string
	HttpEnabled    bool
	HttpPath       string
	HttpPort       int
	StaticDataPath string
	StaticDataPort int
	DbPath         string
	ConfigFile     string
	PipeFilePath   string
}

var env Environment

func main() {
	var (
		grpcServer controller.Server
	)

	parseConfig()

	// set configuration for grpc server
	grpcServer.EnableTls = env.GrpcEnableTls
	grpcServer.ListenPort = env.GrpcListenPort
	grpcServer.TlsCert = env.GrpcTlsCert
	grpcServer.TlsKey = env.GrpcTlsKey
	grpcServer.ConfigFile = env.ConfigFile
	grpcServer.DbPath = env.DbPath
	grpcServer.StaticDataPath = env.StaticDataPath
	grpcServer.StaticDataPort = env.StaticDataPort
	grpcServer.PipeFilePath = env.PipeFilePath

	go grpcServer.Start()

	if env.HttpEnabled {
		log.Printf("tracker-controller spa path: %s\n", env.HttpPath)
		log.Printf("tracker-controller spa listening on port: %d\n", env.HttpPort)
		go controller.HttpServeSPA(fmt.Sprintf(":%d", env.HttpPort), env.HttpPath)

	} else {
		log.Println("tracker-controller disabled")
	}

	for {
		time.Sleep(30 * time.Second)
	}
}

func parseConfig() {

	/// Name of the command line arguments we accept
	paramHttpEnabled := "http-enabled"
	paramHttpPath := "http-path"
	paramHttpPort := "http-port"
	paramGrpcListenPort := "grpc-listen-port"
	paramStaticDataPath := "static-file-path"
	paramStaticDataPort := "static-file-port"
	paramConfigFile := "config-file"
	paramDbPath := "db-path"
	paramPipePath := "pipe-path"

	/// Pull environment variables
	envHttpEnabled := os.Getenv(paramHttpEnabled)
	envHttpPath := os.Getenv(paramHttpPath)
	envFontendPort := os.Getenv(paramHttpPort)
	envGrpcListenPort := os.Getenv(paramGrpcListenPort)

	envStaticDataPath := os.Getenv(paramHttpPath)
	envStaticDataPort := os.Getenv(paramHttpPort)

	envDbPath := os.Getenv(paramDbPath)
	envPipePath := os.Getenv(paramPipePath)
	envConfigFile := os.Getenv(paramConfigFile)

	/// Check for commandline variables
	flag.BoolVar(&env.HttpEnabled, paramHttpEnabled, true, "Enable / Disable the http frontend")
	flag.StringVar(&env.HttpPath, paramHttpPath, "/app/frontend", "Path to the frontend")
	flag.IntVar(&env.HttpPort, paramHttpPort, 8080, "Port for the front end UI")
	flag.IntVar(&env.GrpcListenPort, paramGrpcListenPort, 8088, "Port for the for the Grpc used by the front end UI")

	flag.StringVar(&env.StaticDataPath, paramStaticDataPath, "/skyhub/data/", "Path to the data folder")
	flag.IntVar(&env.StaticDataPort, paramStaticDataPort, 3000, "Port for the data folder http server")
	flag.StringVar(&env.DbPath, paramDbPath, "/skyhub/db/tracker.db", "Path to sqlite db")
	flag.StringVar(&env.PipeFilePath, paramPipePath, "/tmp/skyhub.pipe", "Path to named pipe file")
	flag.StringVar(&env.ConfigFile, paramConfigFile, "/skyhub/etc/tracker.yml", "path to config file")

	flag.Parse()

	/// Override the command line args with env variables
	if len(envHttpEnabled) > 0 {
		tempHttpEnabled, err0 := strconv.ParseBool(envHttpEnabled)
		if err0 == nil {
			env.HttpEnabled = tempHttpEnabled
		}
	}
	if len(envHttpPath) > 0 {
		env.HttpPath = envHttpPath
	}
	if len(envFontendPort) > 0 {
		tempHttpPort, err1 := strconv.ParseInt(envFontendPort, 10, 32)
		if err1 == nil {
			env.HttpPort = int(tempHttpPort)
		}
	}
	if len(envGrpcListenPort) > 0 {
		tempGrpcListenPort, err2 := strconv.ParseInt(envGrpcListenPort, 10, 32)
		if err2 == nil {
			env.GrpcListenPort = int(tempGrpcListenPort)
		}
	}
	if len(envStaticDataPath) > 0 {
		env.StaticDataPath = envStaticDataPath
	}
	if len(envStaticDataPort) > 0 {
		tempStaticDataPort, err3 := strconv.ParseInt(envStaticDataPort, 10, 32)
		if err3 == nil {
			env.StaticDataPort = int(tempStaticDataPort)
		}
	}
	if len(envDbPath) > 0 {
		env.DbPath = envDbPath
	}

	if len(envPipePath) > 0 {
		env.PipeFilePath = envPipePath
	}

	if len(envConfigFile) > 0 {
		env.ConfigFile = envConfigFile
	}
}
