package datacollector

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nycdavid/minutes/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type (
	Mod func(*dataCollector)

	dataCollector struct {
		hbchan         chan Heartbeat
		buckets        map[string][]Heartbeat
		db             *gorm.DB
		sqliteFilePath string
	}

	Heartbeat struct {
		App       string
		Timestamp time.Time
	}
)

func WithBuckets(buckets map[string][]Heartbeat) Mod {
	return func(d *dataCollector) {
		d.buckets = buckets
	}
}

func WithHeartbeatChannel(hbchan chan Heartbeat) Mod {
	return func(dataCollector *dataCollector) {
		dataCollector.hbchan = hbchan
	}
}

func WithSQLiteFilePath(filePath string) Mod {
	return func(dataCollector *dataCollector) {
		dataCollector.sqliteFilePath = filePath
	}
}

func New(opts ...Mod) *dataCollector {
	dc := &dataCollector{
		buckets: make(map[string][]Heartbeat),
	}

	for _, opt := range opts {
		opt(dc)
	}

	if dc.hbchan == nil {
		dc.hbchan = make(chan Heartbeat)
	}

	if _, err := os.Stat(dc.sqliteFilePath); err != nil {
		panic(fmt.Sprintf("error finding db file: %v", err))
	}

	db, err := gorm.Open(sqlite.Open(dc.sqliteFilePath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %v", err))
	}

	dc.db = db

	return dc
}

func (d *dataCollector) Run() {
	for {
		select {
		case hb, ok := <-d.hbchan:
			if !ok {
				return
			}

			d.buckets[hb.App] = append(d.buckets[hb.App], hb)
		}
	}
}

func (d *dataCollector) Flush() {
	for app, hbs := range d.buckets {
		for _, hb := range hbs {
			session := &models.Session{
				Application:    hb.App,
				StartTimestamp: time.Now().UTC().UnixMilli(),
			}
			if err := gorm.G[models.Session](d.db).Create(context.Background(), session); err != nil {
				panic(fmt.Sprintf("error creating session: %v", err))
			}
			// convert to session
			// persist to disk
		}

		// empty bucket
		delete(d.buckets, app)
	}
}

func (d *dataCollector) Buckets() map[string][]Heartbeat {
	return d.buckets
}

func (d *dataCollector) HeartbeatChannel() chan Heartbeat {
	return d.hbchan
}
