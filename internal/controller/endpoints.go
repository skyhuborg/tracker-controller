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
package controller

import (
	"context"
	"fmt"
	_ "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	pb "gitlab.com/uaptn/proto-tracker-controller-go"
	"gitlab.com/uaptn/uaptn/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
    Handle     *grpc.Server
    ListenPort int
    EnableTls  bool
    TlsKey     string
    TlsCert    string
	ConfigFile string


	config *common.Config
	dbPath string
    db    *common.DB
}

func (s *Server) OpenConfig() err {
	err := s.config.Open(s.ConfigFile)

	if err != nil {
		log.Printf("Error: failed opening configuration: %s", s.ConfigFile)
		return err
	}
	return nil
}

func (s *Server) ConnectDb() *gorm.DB {
	db := common.DB{}

	err = db.Open(s.dbPath)

	if err != nil {
		grpclog.Printf("Error: %s\n")
		return r, err
	}
}

func (s *Server) Close() {
	s.config.Close()
	s.db.Close()
}




func StartGrpc(port int, dbPath string) {
	//lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	//if err != nil {
	//    log.Fatalf("failed to listen: %v", err)
	//}

	s := grpc.NewServer()

	pb.RegisterControllerServer(s, &server{dbPath: dbPath})

	grpclog.SetLogger(log.New(os.Stdout, "uaptn-controller: ", log.LstdFlags))

	wrappedServer := grpcweb.WrapServer(s,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}))

	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}

	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(handler),
	}

	grpclog.Printf("Starting server without TLS on port %d", port)
	if err := httpServer.ListenAndServe(); err != nil {
		grpclog.Fatalf("failed starting http server: %v", err)
	}
}

func (s *server) SetConfig(ctx context.Context, in *pb.SetConfigReq) (*pb.SetConfigResp, error) {
	r := pb.SetConfigResp{}
	config := common.Config{}

	err := config.Open("/uaptn/etc/node.ini")

	if err != nil {
		log.Printf("Error: failed opening configuration")
	}

	defer config.Close()

	grpclog.Printf("SetConfig called\n")
	config.SetConfigFromPb(in.Config)

	config.Save()

	return &r, nil
}

func (s *server) GetIsConfigured(ctx context.Context, in *pb.GetIsConfiguredReq) (*pb.GetIsConfiguredResp, error) {
	r := pb.GetIsConfiguredResp{}
	config := common.Config{}

	err := config.Open("/uaptn/etc/node.ini")

	if err != nil {
		log.Printf("Error: failed opening configuration")
	}

	defer config.Close()

	grpclog.Printf("GetIsConfigured called\n")

	r.IsConfigured = config.GetIsConfigured()

	return &r, nil
}

func (s *server) GetConfig(ctx context.Context, in *pb.GetConfigReq) (*pb.GetConfigResp, error) {
	r := pb.GetConfigResp{}

	config := common.Config{}

	err := config.Open("/uaptn/etc/node.ini")

	if err != nil {
		log.Printf("Error: failed opening configuration")
		return &r, err
	}

	defer config.Close()

	grpclog.Printf("GetConfig called\n")
	r.Config = config.GetConfigPb()
	return &r, nil
}

func (s *server) GetEvents(ctx context.Context, in *pb.GetEventsReq) (*pb.GetEventsResp, error) {
	r := &pb.GetEventsResp{}
	var err error
	db := common.DB{}

	err = db.Open(s.dbPath)

	if err != nil {
		grpclog.Printf("Error: %s\n")
		return r, err
	}

	defer db.Close()

	events, total, err := db.GetEvents(in.Limit)

	for _, e := range events {
		ts, _ := ptypes.TimestampProto(e.CreatedAt)
		r.Event = append(r.Event, &pb.Event{Id: e.Id, CreatedAt: ts, Type: e.Type, Source: e.Source, Sensor: e.Sensor, Duration: e.Duration})
	}

	f_total := float64(total)
	f_limit := float64(in.Limit)
	n_pages := int32(math.Ceil(f_total / f_limit))

	r.Npages = n_pages
	r.Page = in.Page
	r.Total = total

	return r, err
}

func (s *server) GetVideoEvents(ctx context.Context, in *pb.GetVideoEventsReq) (*pb.GetVideoEventsResp, error) {
	r := &pb.GetVideoEventsResp{}
	var err error
	db := common.DB{}

	err = db.Open(s.dbPath)

	if err != nil {
		grpclog.Printf("Error: %s\n")
		return r, err
	}

	defer db.Close()

	events, total, err := db.GetVideoEvents(in.Limit)

	for _, e := range events {
		ts, _ := ptypes.TimestampProto(e.CreatedAt)
		base := filepath.Base(e.Uri)
		uri := fmt.Sprintf("http://localhost:3000/video/%s", base)
		base = filepath.Base(e.Thumbnail)
		thumb := fmt.Sprintf("http://localhost:3000/thumb/%s", base)
		r.Video = append(r.Video, &pb.VideoEvent{EventId: e.EventId, CreatedAt: ts, Uri: uri, Thumb: thumb})
	}

	f_total := float64(total)
	f_limit := float64(in.Limit)
	n_pages := int32(math.Ceil(f_total / f_limit))

	r.Npages = n_pages
	r.Page = in.Page
	r.Total = total

	return r, err
}
