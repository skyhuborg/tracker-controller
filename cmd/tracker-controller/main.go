/*
MIT License
-----------

Copyright (c) 2020 Steve McDaniel

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
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gitlab.com/uaptn/uaptn/internal/controller"
)

type Environment struct {
    GrpcListenPort   int
    GrpcEnableTls    bool
    GrpcTlsKey       string
    GrpcTlsCert      string
	FrontendEnabled  bool
	FrontendPath     string
	FrontendPort     int
	FrontendGRPCPort int
	StaticDataPath   string
	StaticDataPort   int
	DbPath           string
	ConfigFile       string
}

var env Environment

func ServeStatic() {
	fs := http.FileServer(http.Dir(env.StaticDataPath))
	http.Handle("/", fs)
	log.Println("tracker-controller static path:" + strconv.Itoa(env.StaticDataPort) + "...")
	log.Println("tracker-controller static listening on port:" + strconv.Itoa(env.StaticDataPort) + "...")
	http.ListenAndServe(":"+strconv.Itoa(env.StaticDataPort), nil)
}

func main() {
    var (
        grpcServer         controller.Server
    )

	parseConfig()

   // set configuration for grpc server
    grpcServer.EnableTls = env.GrpcEnableTls
    grpcServer.ListenPort = env.GrpcListenPort
    grpcServer.TlsCert = env.GrpcTlsCert
    grpcServer.TlsKey = env.GrpcTlsKey
    grpcServer.ConfigFile = env.ConfigFile
    grpcServer.DbPath = env.DbPath

	go grpcServer.Start()

	if env.FrontendEnabled {
		log.Println("tracker-controller spa path:" + env.FrontendPath)
		log.Println("tracker-controller spa listening on port:" + strconv.Itoa(env.FrontendPort) + "...")
		go controller.HttpServeSPA(":"+strconv.Itoa(env.FrontendPort), env.FrontendPath)

	} else {
		log.Println("tracker-controller disabled")
	}
	go ServeStatic()

	for {
		time.Sleep(30 * time.Second)
	}
}

func parseConfig() {

	/// Name of the command line arguments we accept
	paramFrontendEnabled := "frontend-enabled"
	paramFrontendPath := "frontend-path"
	paramFrontendPort := "frontend-port"
	paramFrontendGrpcPort := "frontend-grpc-port"
	paramStaticDataPath := "static-file-path"
	paramStaticDataPort := "static-file-port"
	paramConfigFile := "config-file"
	paramDbPath := "db-path"

	/// Pull environment variables
	envFrontendEnabled := os.Getenv(paramFrontendEnabled)
	envFrontendPath := os.Getenv(paramFrontendPath)
	envFontendPort := os.Getenv(paramFrontendPort)
	envFrontEndGrpcPort := os.Getenv(paramFrontendGrpcPort)

	envStaticDataPath := os.Getenv(paramFrontendPath)
	envStaticDataPort := os.Getenv(paramFrontendPort)

	envDbPath := os.Getenv(paramDbPath)
	envConfigFile := os.Getenv(paramConfigFile)

	/// Check for commandline variables
	flag.BoolVar(&env.FrontendEnabled, paramFrontendEnabled, true, "Enable / Disable the http frontend")
	flag.StringVar(&env.FrontendPath, paramFrontendPath, "/app/frontend", "Path to the frontend")
	flag.IntVar(&env.FrontendPort, paramFrontendPort, 8080, "Port for the front end UI")
	flag.IntVar(&env.FrontendGRPCPort, paramFrontendGrpcPort, 8088, "Port for the for the GRPC used by the front end UI")

	flag.StringVar(&env.StaticDataPath, paramStaticDataPath, "/uaptn/data/", "Path to the data folder")
	flag.IntVar(&env.StaticDataPort, paramStaticDataPort, 3000, "Port for the data folder http server")
	flag.StringVar(&env.DbPath, paramDbPath, "/uaptn/db/tracker.db", "Path to sqlite db")
	flag.StringVar(&env.ConfigFile, paramConfigFile, "/uaptn/etc/tracker.yml", "path to config file")

	flag.Parse()

	/// Override the command line args with env variables
	if len(envFrontendEnabled) > 0 {
		tempFrontendEnabled, err0 := strconv.ParseBool(envFrontendEnabled)
		if err0 == nil {
			env.FrontendEnabled = tempFrontendEnabled
		}
	}
	if len(envFrontendPath) > 0 {
		env.FrontendPath = envFrontendPath
	}
	if len(envFontendPort) > 0 {
		tempFrontendPort, err1 := strconv.ParseInt(envFontendPort, 10, 32)
		if err1 == nil {
			env.FrontendPort = int(tempFrontendPort)
		}
	}
	if len(envFrontEndGrpcPort) > 0 {
		tempFrontendGrpcPort, err2 := strconv.ParseInt(envFrontEndGrpcPort, 10, 32)
		if err2 == nil {
			env.FrontendGRPCPort = int(tempFrontendGrpcPort)
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

    if len(envConfigFile) > 0 {
        env.ConfigFile = envConfigFile
    }
}
