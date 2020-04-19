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
package common

import (
	"errors"
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	pb "gitlab.com/uaptn/proto-tracker-controller-go"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Camera struct {
	Name     string `yaml:"name"`
	Enabled  bool   `yaml:"enabled"`
	Location string `yaml:"location"`
	Protocol string `yaml:"protocol"`
	Ip       string `yaml:"ip"`
	Port     int32  `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Storage struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
}

type Settings struct {
	Configured bool   `yaml:"configured"`
	Uuid       string `yaml:"uuid"`
	NodeName   string `yaml:"nodename"`
	Hostname   string `yaml:"hostname"`
	Camera     []Camera
	Storage    []Storage
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func touchFile(filename string) (success bool) {
	var (
		fileDir string
		err     error
	)

	// extract the dir of the filename
	fileDir = filepath.Dir(filename)

	// create entire directory structure
	os.MkdirAll(fileDir, os.ModePerm)

	// Create the file
	f, err := os.Create(filename)

	if err != nil {
		return false
	}

	f.Close()

	return true
}

type Config struct {
	NodeName string
	uri      string
	s        Settings
}

func (c *Config) GetCameras() []Camera {
	return c.s.Camera
}

func (c *Config) GetConfigPb() *pb.Config {
	config := pb.Config{}

	config.Uuid = c.s.Uuid
	config.Hostname = c.s.Hostname
	config.NodeName = c.s.NodeName
	config.Configured = c.s.Configured

	for _, c := range c.s.Camera {
		config.Camera = append(config.Camera, &pb.CameraConfig{
			Name:     c.Name,
			Enabled:  c.Enabled,
			Location: c.Location,
			Protocol: c.Protocol,
			Ip:       c.Ip,
			Port:     c.Port,
			Username: c.Username,
			Password: c.Password})
	}

	for _, s := range c.s.Storage {
		config.Storage = append(config.Storage, &pb.StorageConfig{
			Name:     s.Name,
			Location: s.Location})
	}

	return &config
}

func (c *Config) SetConfigFromPb(config *pb.Config) {
	s := Settings{}

	for _, camera := range config.Camera {
		s.Camera = append(s.Camera, Camera{
			Name:     camera.Name,
			Enabled:  camera.Enabled,
			Location: camera.Location,
			Protocol: camera.Protocol,
			Ip:       camera.Ip,
			Port:     camera.Port,
			Username: camera.Username,
			Password: camera.Password})
	}

	for _, storage := range config.Storage {
		s.Storage = append(s.Storage, Storage{
			storage.Name,
			storage.Location})
	}

	s.Uuid = config.Uuid
	s.Hostname = config.Hostname
	s.NodeName = config.NodeName
	s.Configured = true

	c.s = s

	d, _ := yaml.Marshal(s)

	ioutil.WriteFile(c.uri, d, 0644)
}

func (c *Config) SetHostname(hostname string) {
	c.s.Hostname = hostname
}

func (c *Config) SetNodeName(nodename string) {
	c.s.NodeName = nodename
}

func (c *Config) SetConfigured(is_configured bool) {
	c.s.Configured = is_configured
}

/* this will create the ini file and
   set default values
*/
func (c *Config) SetDefaults() {
	log.Println("setting defaults")
	uid, _ := uuid.NewUUID()
	c.s.Uuid = uid.String()
	c.s.NodeName = randomdata.SillyName()
	c.s.Configured = false
}

func (c *Config) GetUuid() string {
	return c.s.Uuid
}

func (c *Config) GetHostname() string {
	return c.s.Hostname
}

func (c *Config) GetIsConfigured() bool {
	return c.s.Configured
}

func (c *Config) GetNodeName() string {
	return c.s.NodeName
}

func (c *Config) Save() {
	d, _ := yaml.Marshal(c.s)
	ioutil.WriteFile(c.uri, d, 0644)
}

func (c *Config) Open(uri string) (err error) {
	var (
		setDefaults bool
		ok          bool
		data        []byte
	)

	setDefaults = false

	if fileExists(uri) == false {
		ok = touchFile(uri)

		if ok == false {
			err = errors.New("Failed creating config")
			return
		}
		setDefaults = true
	}

	data, err = ioutil.ReadFile(uri)

	if err != nil {
		log.Println("Readfile failed: %s\n", err)
		return
	}

	err = yaml.Unmarshal(data, &c.s)

	if err != nil {
		log.Println("unmarshal failed: %s\n", err)
		return
	}

	c.uri = uri

	if setDefaults == true {
		c.SetDefaults()
		c.Save()
	}

	return nil
}

func (c *Config) Close() {
}