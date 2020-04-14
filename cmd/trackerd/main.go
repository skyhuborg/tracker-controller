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
	"gitlab.com/uaptn/uaptn/internal/trackerd"
	"gitlab.com/uaptn/uaptn/internal/upload"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

type Environment struct {
	GrpcListenPort   int
	GrpcEnableTls    bool
	GrpcTlsKey       string
	GrpcTlsCert      string
	UploadListenPort int
	UploadEnableTls  bool
	UploadTlsKey     string
	UploadTlsCert    string
	UploadPath       string
	DbName           string
	DbHost           string
	DbPort           int
	DbUser           string
	DbPassword       string
}

var env Environment

func main() {
	var (
		grpcServer         trackerd.Server
		uploadServer       *upload.Server
		uploadServerConfig upload.ServerConfig
		done               chan int
		c                  chan os.Signal
	)

	// load configuration
	parseConfig()

	// set configuration for upload server
	uploadServerConfig.ListenPort = env.UploadListenPort
	uploadServerConfig.ChunkSize = 4096
	uploadServerConfig.UploadPath = env.UploadPath

	// set configuration for grpc server
	grpcServer.EnableTls = env.GrpcEnableTls
	grpcServer.ListenPort = env.GrpcListenPort
	grpcServer.TlsCert = env.GrpcTlsCert
	grpcServer.TlsKey = env.GrpcTlsKey
	grpcServer.DbHost = env.DbHost
	grpcServer.DbName = env.DbName
	grpcServer.DbPort = env.DbPort
	grpcServer.DbUser = env.DbUser
	grpcServer.DbPassword = env.DbPassword

	// create channel for signal and done channel
	done = make(chan int)
	c = make(chan os.Signal)

	// initialize a new uploadServer with config
	uploadServer = upload.NewServer(&uploadServerConfig)

	// setup signal to gracefully shutdown on SIGTERM
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func(uploadServer *upload.Server, grpcServer trackerd.Server) {
		<-c
		log.Println("Shutting down server")
		uploadServer.Close()
		os.Exit(0)
	}(uploadServer, grpcServer)

	// start the upload server
	go uploadServer.Start()

	// start the grpc server
	go grpcServer.Start()

	<-done
}

func parseConfig() {
	/// Name of the command line arguments we accept
	var (
		paramGrpcListenPort   string = "grpc-listen-port"
		paramGrpcEnableTls    string = "grpc-enable-tls"
		paramGrpcTlsKey       string = "grpc-tls-key"
		paramGrpcTlsCert      string = "grpc-tls-cert"
		paramUploadListenPort string = "upload-listen-port"
		paramUploadEnableTls  string = "upload-enable-tls"
		paramUploadTlsKey     string = "upload-tls-key"
		paramUploadTlsCert    string = "upload-tls-cert"
		paramUploadPath       string = "upload-path"
		paramDbPort           string = "db-port"
		paramDbHost           string = "db-host"
		paramDbName           string = "db-name"
		paramDbUser           string = "db-user"
		paramDbPassword       string = "db-password"
	)

	/// Pull environment variables
	envGrpcListenPort := os.Getenv(paramGrpcListenPort)
	envGrpcEnableTls := os.Getenv(paramGrpcEnableTls)
	envGrpcTlsKey := os.Getenv(paramGrpcTlsKey)
	envGrpcTlsCert := os.Getenv(paramGrpcTlsCert)

	envUploadListenPort := os.Getenv(paramUploadListenPort)
	envUploadEnableTls := os.Getenv(paramUploadEnableTls)
	envUploadTlsKey := os.Getenv(paramUploadTlsKey)
	envUploadTlsCert := os.Getenv(paramUploadTlsCert)
	envUploadPath := os.Getenv(paramUploadPath)

	envDbPort := os.Getenv(paramDbPort)
	envDbHost := os.Getenv(paramDbHost)
	envDbName := os.Getenv(paramDbName)
	envDbUser := os.Getenv(paramDbUser)
	envDbPassword := os.Getenv(paramDbPassword)

	/// Check for commandline variables
	flag.IntVar(&env.GrpcListenPort, paramGrpcListenPort, 8088, "Port for the for the gRPC server")
	flag.BoolVar(&env.GrpcEnableTls, paramGrpcEnableTls, false, "Enable TLS on gRPC connection")
	flag.StringVar(&env.GrpcTlsCert, paramGrpcTlsCert, "/app/cert.pem", "Path to TLS certificate")
	flag.StringVar(&env.GrpcTlsKey, paramGrpcTlsKey, "/app/privkey.pem", "Path to TLS Private Key")

	flag.IntVar(&env.UploadListenPort, paramUploadListenPort, 8090, "Port for the for the File Upload server")
	flag.BoolVar(&env.UploadEnableTls, paramUploadEnableTls, false, "Enable TLS on File Upload connection")
	flag.StringVar(&env.UploadTlsCert, paramUploadTlsCert, "/app/cert.pem", "Path to TLS certificate")
	flag.StringVar(&env.UploadTlsKey, paramUploadTlsKey, "/app/privkey.pem", "Path to TLS Private Key")
	flag.StringVar(&env.UploadPath, paramUploadPath, "/app/data/", "Root path where files will be uploaded")

	flag.StringVar(&env.DbHost, paramDbHost, "localhost", "Hostname running the db")
	flag.StringVar(&env.DbName, paramDbName, "uaptn", "Name of the database")
	flag.IntVar(&env.DbPort, paramDbPort, 3306, "Port to database")
	flag.StringVar(&env.DbUser, paramDbUser, "root", "username for database")
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

	if len(envUploadListenPort) > 0 {
		tempUploadListenPort, err2 := strconv.ParseInt(envUploadListenPort, 10, 32)
		if err2 == nil {
			env.UploadListenPort = int(tempUploadListenPort)
		}
	}

	if len(envUploadEnableTls) > 0 {
		tempUploadEnableTls, err2 := strconv.ParseBool(envUploadEnableTls)
		if err2 == nil {
			env.UploadEnableTls = bool(tempUploadEnableTls)
		}
	}

	if len(envUploadTlsCert) > 0 {
		env.UploadTlsCert = envUploadTlsCert
	}

	if len(envUploadTlsKey) > 0 {
		env.UploadTlsKey = envUploadTlsKey
	}

	if len(envUploadPath) > 0 {
		env.UploadPath = envUploadPath
	}

	if len(envDbHost) > 0 {
		env.DbHost = envDbHost
	}

	if len(envDbName) > 0 {
		env.DbName = envDbName
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
