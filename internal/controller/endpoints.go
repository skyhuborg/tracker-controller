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
	"gitlab.com/uaptn/tracker-controller/internal/common"
	"gitlab.com/uaptn/trackerdb"
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
	DbPath     string
	StaticDataPath string
	StaticDataPort int


	config common.Config
	db     trackerdb.DB
}

func (s *Server) OpenConfig() (err error) {
	err = s.config.Open(s.ConfigFile)

	if err != nil {
		log.Printf("Error: failed opening configuration: %s", s.ConfigFile)
		return err
	}
	return nil
}

func (s *Server) ConnectDb() error {
	db := trackerdb.DB{}

	err := db.Open(s.DbPath)

	if err != nil {
		grpclog.Printf("Error: %s\n")
		return err
	}

	s.db = db
	return nil
}

func (s *Server) StartFileServer() error {
	log.Printf("Service '%s' on :%d\n", s.StaticDataPath, s.StaticDataPort)
	fs := http.FileServer(http.Dir(s.StaticDataPath))
	http.Handle("/", fs)

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.StaticDataPort), nil)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (s *Server) Close() {
	s.config.Close()
	s.db.Close()
}

func (s *Server) Start() {
	var (
		err error
	)

	err = s.OpenConfig()

	if err != nil {
		log.Printf("OpenConfig failed: %s\n", err)
		return
	}

	err = s.ConnectDb()

	if err != nil {
		log.Printf("ConnectDb failed with: %s\n", err)
		return
	}


	go s.StartFileServer()

	s.Handle = grpc.NewServer()

	pb.RegisterControllerServer(s.Handle, s)

	grpclog.SetLogger(log.New(os.Stdout, "tracker-controller: ", log.LstdFlags))

	wrappedServer := grpcweb.WrapServer(s.Handle,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return true
		}))

	handler := func(resp http.ResponseWriter, req *http.Request) {
		wrappedServer.ServeHTTP(resp, req)
	}

	httpServer := http.Server{
		Addr:    fmt.Sprintf(":%d", s.ListenPort),
		Handler: http.HandlerFunc(handler),
	}

	grpclog.Printf("Starting server without TLS on port %d", s.ListenPort)
	if err := httpServer.ListenAndServe(); err != nil {
		grpclog.Fatalf("failed starting http server: %v", err)
	}
}

func (s *Server) SetConfig(ctx context.Context, in *pb.SetConfigReq) (*pb.SetConfigResp, error) {
	r := pb.SetConfigResp{}

	grpclog.Printf("SetConfig called\n")
	s.config.SetConfigFromPb(in.Config)
	s.config.Save()

	return &r, nil
}

func (s *Server) GetIsConfigured(ctx context.Context, in *pb.GetIsConfiguredReq) (*pb.GetIsConfiguredResp, error) {
	r := pb.GetIsConfiguredResp{}

	grpclog.Printf("GetIsConfigured called\n")

	r.IsConfigured = s.config.GetIsConfigured()

	return &r, nil
}

func (s *Server) GetConfig(ctx context.Context, in *pb.GetConfigReq) (*pb.GetConfigResp, error) {
	r := pb.GetConfigResp{}

	grpclog.Printf("GetConfig called\n")

	r.Config = s.config.GetConfigPb()

	return &r, nil
}

func (s *Server) GetEvents(ctx context.Context, in *pb.GetEventsReq) (*pb.GetEventsResp, error) {
	r := &pb.GetEventsResp{}
	var err error

	events, total, err := s.db.GetEvents(in.Limit)

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

func (s *Server) GetVideoEvents(ctx context.Context, in *pb.GetVideoEventsReq) (*pb.GetVideoEventsResp, error) {
	r := &pb.GetVideoEventsResp{}
	var err error

	log.Println("GetVideoEvents called")
	events, total, err := s.db.GetVideoEvents(in.Limit)

	log.Println(events)
	for _, e := range events {
		ts, _ := ptypes.TimestampProto(e.CreatedAt)
		base := filepath.Base(e.Uri)
		uri := fmt.Sprintf("http://localhost:3000/video/%s", base)
		base = filepath.Base(e.Thumbnail)
		thumb := fmt.Sprintf("http://localhost:3000/thumbnail/%s", base)
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
