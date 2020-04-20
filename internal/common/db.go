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
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"sync"
	"time"
)

type Event struct {
	Id        string    `gorm:"type:text; PRIMARY_KEY"`
	CreatedAt time.Time `gorm:"type:datetime;Column:created_at"`
	StartedAt time.Time `gorm:"type:datetime;Column:started_at"`
	EndedAt   time.Time `gorm:"type:datetime;Column:ended_at"`
	Duration  int64     `gorm:"type:int64"`
	Type      string    `gorm:"type:text"`
	Source    string    `gorm:"type:text"`
	Sensor    string    `gorm:"type:text"`
}

type VideoEvent struct {
	EventId      string    `gorm:"column:event_id; PRIMARY_KEY" json:"event_id"`
	CreatedAt    time.Time `gorm:"type:datetime;Column:created_at"`
	Uri          string    `gorm:"type:text"`
	Thumbnail    string    `gorm:"type:text"`
	IsUploaded   bool      `gorm:"default:false"`
	IsPending    bool      `gorm:"default:false"`
	IsInprogress bool      `gorm:"default:false"`
}

type Sensor struct {
	EventId   string    `gorm:"column:event_id;null;index;default:NULL" json:"event_id"`
	CreatedAt time.Time `gorm:"type:datetime;Column:created_at"`
	Data      []byte    `gorm:"type:blob"`
}

type DB struct {
	h *gorm.DB
}

var mutex = &sync.Mutex{}

func (db *DB) Open(uri string) (err error) {
	mutex.Lock()

	db.h, err = gorm.Open("sqlite3", uri)
	if err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	err = db.h.AutoMigrate(
		&Event{},
		&VideoEvent{},
		&Sensor{}).Error

	if err != nil {
		log.Printf("Migrate failed: %s\n", err)
	}

	log.Printf("Opening database from %s\n", uri)

	db.h.BlockGlobalUpdate(true)

	mutex.Unlock()

	return
}

func (db *DB) Save(v interface{}) {
	db.h.Save(v)
}

func (db *DB) Close() {
	mutex.Lock()
	db.h.Close()
	db = nil
	mutex.Unlock()
}

func (db *DB) StartEvent(uuid string, ts time.Time, eventType string, eventSource string, sensor string) {
	var (
		rec Event
	)

	rec.Id = uuid
	rec.StartedAt = ts
	rec.Type = eventType
	rec.Source = eventSource
	rec.Sensor = sensor

	err := db.h.Create(&rec).Error

	if err != nil {
		log.Printf("Create returned: %s\n", err)
	}
}

func (db *DB) StopEvent(
	uuid string,
	ts time.Time,
	duration int64) {
	var (
		rec Event
	)

	rec.Id = uuid

	err := db.h.Model(&rec).Updates(Event{EndedAt: ts, Duration: duration}).Error

	if err != nil {
		log.Printf("Update returned: %s\n", err)
	}
}

func (db *DB) AddVideoEvent(eventId string, videoUri string, thumbnailUri string) (err error) {
	var (
		rec VideoEvent
	)

	rec.EventId = eventId
	rec.Uri = videoUri
	rec.Thumbnail = thumbnailUri

	err = db.h.Create(&rec).Error

	if err != nil {
		log.Printf("Failed adding VideoEvent: %s\n", err)
	}
	return
}

func (db *DB) AddEvent(rec *Event, event_data interface{}) (err error) {
	db.h.Create(rec)
	switch event_data.(type) {
	case *VideoEvent:
		ed, ok := event_data.(*VideoEvent)

		if ok == true {
			ed.EventId = rec.Id
			db.h.Create(event_data)
		}
	default:
		log.Println("unknown")
	}

	return
}

func (db *DB) GetEvents(limit int32) (events []Event, count int32, err error) {
	db.h.Table("events").Count(&count)
	db.h.Limit(limit).Find(&events)

	return
}

func (db *DB) GetVideoEvents(limit int32) (events []VideoEvent, count int32, err error) {
	db.h.Table("video_events").Count(&count)
	log.Println(db.h.Limit(limit).Find(&events).Error)
	log.Println(events)
	return
}

func (db *DB) GetVideoEventNotUploaded() (videoEvent VideoEvent, err error) {
	err = db.h.Where("is_uploaded = ? AND is_pending = ?", false, false).First(&videoEvent).Error
	return
}

func (db *DB) AddSensorData(rec *Sensor) (err error) {
	db.h.Create(rec)
	return nil
}
