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
	"gitlab.com/uaptn/uaptn/internal/portald"
	"os"
	"strconv"
)

type Environment struct {
	GrpcListenPort int
	GrpcEnableTls  bool
	GrpcTlsKey     string
	GrpcTlsCert    string
	StaticDataPath string
	StaticDataPort int
	DbHost         string
	DbPort         int
	DbUser         string
	DbPassword     string
}

var env Environment

func main() {
	parseConfig()
	var done = make(chan int)

	server := ui.Server{}

	server.EnableTls = env.GrpcEnableTls
	server.ListenPort = env.GrpcListenPort
	server.TlsCert = env.GrpcTlsCert
	server.TlsKey = env.GrpcTlsKey
	server.DbHost = env.DbHost
	server.DbPort = env.DbPort
	server.DbUser = env.DbUser
	server.DbPassword = env.DbPassword

	go server.Start()

	<-done
}

func parseConfig() {
	/// Name of the command line arguments we accept
	paramGrpcListenPort := "grpc-listen-port"
	paramGrpcEnableTls := "grpc-enable-tls"
	paramGrpcTlsKey := "grpc-tls-key"
	paramGrpcTlsCert := "grpc-tls-cert"
	paramStaticDataPath := "static-data-path"
	paramStaticDataPort := "static-data-port"
	paramDbPort := "db-port"
	paramDbHost := "db-host"
	paramDbUser := "db-user"
	paramDbPassword := "db-password"

	/// Pull environment variables
	envGrpcListenPort := os.Getenv(paramGrpcListenPort)
	envGrpcEnableTls := os.Getenv(paramGrpcEnableTls)
	envGrpcTlsKey := os.Getenv(paramGrpcTlsKey)
	envGrpcTlsCert := os.Getenv(paramGrpcTlsCert)

	envStaticDataPath := os.Getenv(paramStaticDataPath)
	envStaticDataPort := os.Getenv(paramStaticDataPort)

	envDbPort := os.Getenv(paramDbPort)
	envDbHost := os.Getenv(paramDbHost)
	envDbUser := os.Getenv(paramDbUser)
	envDbPassword := os.Getenv(paramDbPassword)

	/// Check for commandline variables
	flag.IntVar(&env.GrpcListenPort, paramGrpcListenPort, 8088, "Port for the for the gRPC server")
	flag.BoolVar(&env.GrpcEnableTls, paramGrpcEnableTls, false, "Enable TLS on gRPC port")
	flag.StringVar(&env.GrpcTlsCert, paramGrpcTlsCert, "/app/cert.pem", "Path to TLS certificate")
	flag.StringVar(&env.GrpcTlsKey, paramGrpcTlsKey, "/app/privkey.pem", "Path to TLS Private Key")
	flag.StringVar(&env.StaticDataPath, paramStaticDataPath, "/uaptn/data/", "Path to the data folder")
	flag.IntVar(&env.StaticDataPort, paramStaticDataPort, 3000, "Port for the data folder http server")
	flag.StringVar(&env.DbHost, paramDbHost, "localhost", "Hostname running the db")
	flag.IntVar(&env.DbPort, paramDbPort, 3306, "Port to database")
	flag.StringVar(&env.DbUser, paramDbUser, "admin", "username for database")
	flag.StringVar(&env.DbPassword, paramDbPassword, "password", "password for database")
	flag.Parse()

	/// Override the command line args with env variables
	if len(envGrpcListenPort) > 0 {
		tempGrpcListenPort, err2 := strconv.ParseInt(envGrpcListenPort, 10, 32)
		if err2 == nil {
			env.GrpcListenPort = int(tempGrpcListenPort)
		}
	}

	if len(envGrpcEnableTls) > 0 {
		tempGrpcEnableTls, err2 := strconv.ParseBool(envGrpcEnableTls)
		if err2 == nil {
			env.GrpcEnableTls = bool(tempGrpcEnableTls)
		}
	}

	if len(envGrpcTlsCert) > 0 {
		env.GrpcTlsCert = envGrpcTlsCert
	}
	if len(envGrpcTlsKey) > 0 {
		env.GrpcTlsKey = envGrpcTlsKey
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

	if len(envDbHost) > 0 {
		env.DbHost = envDbHost
	}

	if len(envDbPort) > 0 {
		tempDbPort, err3 := strconv.ParseInt(envDbPort, 10, 32)
		if err3 == nil {
			env.DbPort = int(tempDbPort)
		}

	}

	if len(envDbUser) > 0 {
		env.DbUser = envDbUser
	}

	if len(envDbPassword) > 0 {
		env.DbPassword = envDbPassword
	}
}
