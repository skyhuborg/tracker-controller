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
	//"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/proto"
	pb "gitlab.com/uaptn/proto-trackerd-go"
	"gitlab.com/uaptn/uaptn/internal/common"
	"gitlab.com/uaptn/uaptn/internal/event"
	"gitlab.com/uaptn/uaptn/internal/sensor"
	"gitlab.com/uaptn/uaptn/internal/tracker"
	"gitlab.com/uaptn/uaptn/internal/upload"
	"gitlab.com/uaptn/uaptn/internal/video"
)

type Environment struct {
	Debug        bool
	DbPath       string
	DataPath     string
	ConfigFile   string
	TrackerdAddr string
	UploadAddr   string
	GpsAddr     string
}

type Tracker struct {
	// Id of Tracker instance
	id string
	// Human readable name of node
	node_name string
	// Command line and Environment variables
	env Environment
	// Pubsub bus where all messages are passed
	bus event.Bus
	// GRPC Client handle for the trackerd
	client tracker.ClientGRPC
	// GRPC Client configuration
	config tracker.ClientGRPCConfig
	// GRPC Client handle for the trackerd
	uploadClient upload.Client
	// GRPC Client configuration
	uploadClientConfig upload.ClientConfig
	// EventMonitor that monitors all events on the bus
	eventMonitor  *event.Monitor
	sensorMonitor sensor.Monitor
	// Configuration settings for the tracker
	settings common.Config
	// database handle
	db common.DB
}

var env Environment
var version string

func InitializeStore() {
	videoDir := fmt.Sprintf("%s/video", env.DataPath)
	thumbnailDir := fmt.Sprintf("%s/thumbnail", env.DataPath)
	etcDir := fmt.Sprintf("%s/etc", env.DataPath)

	os.MkdirAll(etcDir, 755)
	os.MkdirAll(videoDir, 755)
	os.MkdirAll(thumbnailDir, 755)
}

func DbOutputThread(ch chan *pb.SensorReport) {
	var (
		db     common.DB
		rec    common.Sensor
		err    error
		report *pb.SensorReport
		data   []byte
	)

	log.Printf("Started DB Video Logging thread\n")

	err = db.Open(env.DbPath)

	if err != nil {
		return
	}

	defer db.Close()

	for {
		report = <-ch

		data, err = proto.Marshal(report)

		if err != nil {
			log.Printf("Failed Marshaling SensorReport: %s\n", err)
			continue
		}
		rec.EventId = report.EventId
		rec.Data = data
		db.AddSensorData(&rec)
	}
}

func DbOutputVideoThread(ch chan video.VideoRecording) {
	var (
		db             common.DB
		videoRecording video.VideoRecording
	)

	log.Printf("Started DB Logging thread\n")
	err := db.Open(env.DbPath)

	if err != nil {
		return
	}

	defer db.Close()

	for {
		videoRecording = <-ch

		db.AddVideoEvent(videoRecording.EventId, videoRecording.Uri, videoRecording.Thumbnail)
	}
}

func VideoUploadThread(t *Tracker) {
	var (
		videoEvent common.VideoEvent
		err        error
	)

	for {
		videoEvent, err = t.db.GetVideoEventNotUploaded()

		if err != nil {
			goto timeout
		}

		videoEvent.IsPending = true
		t.db.Save(videoEvent)

		err = t.uploadClient.Upload(videoEvent.EventId, videoEvent.Uri, videoEvent.Thumbnail)

		if err != nil {
			log.Printf("Error: Upload failed  %s\n", err)
			log.Println(videoEvent.Uri)
			videoEvent.IsPending = false
			t.db.Save(videoEvent)
			goto timeout
		}
		videoEvent.IsUploaded = true
		t.db.Save(videoEvent)

		continue
	timeout:
		time.Sleep(5 * time.Second)
	}
}

func TrackerOutputThread(client tracker.ClientGRPC, ch chan *pb.SensorReport) {
	/*
		client := pb.NewTrackerdClient(conn)
		for {
			report := <-ch
			_, err := client.AddSensor(context.Background(), report)

			if err != nil {
				log.Printf("Error: %s\n", err)
			}
		}
	*/
}

func parseArgs() {
	paramDbPath := "db-path"
	paramDataPath := "data-path"
	paramConfigFile := "config-file"
	paramTrackerdAddr := "trackerd-server-addr"
	paramUploadAddr := "upload-server-addr"
	paramGpsAddr := "gpsd-addr"
	paramDebug := "debug"

	envDbPath := os.Getenv(paramDbPath)
	envDataPath := os.Getenv(paramDbPath)
	envConfigFile := os.Getenv(paramConfigFile)
	envTrackerdAddr := os.Getenv(paramTrackerdAddr)
	envUploadAddr := os.Getenv(paramUploadAddr)
	envGpsAddr := os.Getenv(paramGpsAddr)
	envDebug := os.Getenv(paramDebug)

	flag.StringVar(&env.DbPath, paramDbPath, "/uaptn/db/tracker.db", "path to database file")
	flag.StringVar(&env.DataPath, paramDataPath, "/uaptn/data", "path to data directory")
	flag.StringVar(&env.ConfigFile, paramConfigFile, "/uaptn/etc/tracker.yml", "path to config file")
	flag.StringVar(&env.TrackerdAddr, paramTrackerdAddr, "localhost:8088", "Trackerd server address")
	flag.StringVar(&env.UploadAddr, paramUploadAddr, "localhost:8090", "Upload server address")
	flag.StringVar(&env.GpsAddr, paramGpsAddr, "localhost:2947", "GPSD server address")
	flag.BoolVar(&env.Debug, paramDebug, false, "Enable debugging")
	flag.Parse()

	if len(envDbPath) > 0 {
		env.DbPath = envDbPath
	}

	if len(envDataPath) > 0 {
		env.DataPath = envDataPath
	}

	if len(envConfigFile) > 0 {
		env.ConfigFile = envConfigFile
	}

	if len(envTrackerdAddr) > 0 {
		env.TrackerdAddr = envTrackerdAddr
	}

	if len(envUploadAddr) > 0 {
		env.UploadAddr = envUploadAddr
	}

	if len(envGpsAddr) > 0 {
		env.GpsAddr = envGpsAddr
	}

	if len(envDebug) > 0 {
		tempDebug, err2 := strconv.ParseBool(envDebug)
		if err2 == nil {
			env.Debug = bool(tempDebug)
		}
	}
}

func NewTracker(env Environment) (t Tracker, err error) {
	var (
		// handle for the gpsd sensor
		gpsdSensor sensor.GPSDSensor
		// handle for the witmotion hw901/jy901 sensor
		jy901Sensor sensor.JY901Sensor
	)

	t.env = env

	err = t.db.Open(env.DbPath)

	if err != nil {
		return
	}

	// Initialize the subscriber map
	t.bus = *event.NewBus()

	// create EventMonitor that is attached to the bus
	t.eventMonitor = event.NewMonitor(&t.bus)

	// Load the local configuration for the tracker
	err = t.settings.Open(env.ConfigFile)

	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	defer t.settings.Close()

	t.id = t.settings.GetUuid()

	if err != nil {
		log.Fatalf("Error: Failed getting Trackerd Id")
		return
	}

	t.node_name = t.settings.GetNodeName()

	if err != nil {
		log.Fatalf("Error: failed getting NodeName")
		return
	}

	// configure and connect to trackerd
	clientConfig := tracker.ClientGRPCConfig{
		Address: env.TrackerdAddr}
	t.client, err = tracker.NewClientGRPC(clientConfig)

	if err != nil {
		log.Fatalf("Error: failed connecting to trackerd: %s\n", err)
		return
	}

	t.uploadClientConfig = upload.ClientConfig{
		ServerAddr: env.UploadAddr,
		ChunkSize:  4096,
	}

	t.uploadClient = upload.NewClient(&t.uploadClientConfig)

	err = t.uploadClient.Connect()

	if err != nil {
		log.Printf("Start() failed: %s\n", err)
		return
	}

	// create new SensorMonitor attached to the bus
	t.sensorMonitor, err = sensor.NewMonitor(&t.bus)

	if err != nil {
		log.Fatalf("Error: failed creating sensor.Monitor: %s", err)
		return
	}

	// poll sensors every second, this will be changed to something different
	t.sensorMonitor.SetPollInterval(time.Second * 1)

	/* register output channels */
	dbChannel := make(chan *pb.SensorReport, 10)
	//trackerdChannel := make(chan *pb.SensorReport, 10)
	videoChannel := make(chan video.VideoRecording, 10)

	t.sensorMonitor.RegisterChannel(dbChannel)
	//t.sensorMonitor.RegisterChannel(trackerdChannel)

	//go TrackerOutputThread(t.client, trackerdChannel)
	go DbOutputThread(dbChannel)
	go DbOutputVideoThread(videoChannel)

	/* register sensors */
	gpsdSensor.Addr = env.GpsAddr
	t.sensorMonitor.Register(&gpsdSensor)
	t.sensorMonitor.Register(&jy901Sensor)

	log.Printf("SensorMonitor:  Started sensor polling")

	go t.sensorMonitor.Poll()
	go t.Start()

	if len(t.settings.GetCameras()) > 0 {
		for _, c := range t.settings.GetCameras() {
			var (
				v         video.Video
				protocol  string
				cameraUri string
			)
			
			if c.Enabled == false {
				continue
			}

			protocol = strings.ToLower(c.Protocol)

			cameraUri = fmt.Sprintf("%s://%s:%s@%s:%d", protocol, c.Username, c.Password, c.Ip, c.Port)

			v.SetBus(&t.bus)

			err = v.Open(cameraUri)

			if err != nil {
				log.Printf("Failed opening device: %s\n", err)
				return
			}
			//defer v.Close()

			motionDetector := video.NewMotionDetector()
			motionDetector.SetMinimumArea(2000)
			motionDetector.EnableBoxing(true)
			motionDetector.SetBus(&t.bus)
			v.RegisterVideoOutput(motionDetector)

			recorder := video.NewRecord()
			recorder.SetBus(&t.bus)
			recorder.SetDataPath(env.DataPath)
			recorder.Register(videoChannel)

			v.RegisterVideoOutput(recorder)
			go v.Start()
		}
	} else {
		log.Println("No cameras detected.  Operating in Sensor only mode.")
	}

	return
}

func (t *Tracker) Stop() {

	t.client.Close()
	t.db.Close()
}

func (t *Tracker) Start() (bSuccess bool) {
	/* Start the event monitor */
	t.eventMonitor.Start()

	go t.Poll()

	go VideoUploadThread(t)

	log.Printf("Node=%s UUID=%s is Online\n", t.node_name, t.id)

	return
}

func (t *Tracker) Poll() {
	systemState := make(chan event.DataEvent, 100)
	t.bus.Subscribe(event.EventStart, systemState)
	t.bus.Subscribe(event.EventStop, systemState)
	db := common.DB{}

	err := db.Open(t.env.DbPath)
	log.Printf("DbPath: %s\n", t.env.DbPath)

	if err != nil {
		log.Printf("DB Open failed: %s\n", err)
		return
	}

	for {
		d := <-systemState
		e, ok := d.Data.(*event.Event)

		if !ok {
			continue
		}

		if d.Topic == event.EventStart {
			db.StartEvent(e.Id, e.StartTime, e.Type.String(), e.Source.Name, e.Source.Sensor)
		} else if d.Topic == event.EventStop {
			db.StopEvent(e.Id, e.EndTime, e.GetDuration())
		}
	}
}

func main() {
	var (
		err     error
		tracker Tracker
	)

	// load all command line args and env variables
	parseArgs()

	// creates the directory structure for the data path
	InitializeStore()

	// create our tracker instance and expose the Environment to it
	tracker, err = NewTracker(env)

	if err != nil {
		log.Printf("Tracker initialization failed: %s\n", err)
		return
	}

	for {
		time.Sleep(30 * time.Second)
	}

	tracker.Stop()
}
