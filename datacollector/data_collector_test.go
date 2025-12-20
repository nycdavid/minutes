package datacollector_test

import (
	"context"
	"testing"
	"time"

	"github.com/nycdavid/minutes/datacollector"
	"github.com/nycdavid/minutes/models"
	_assert "github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Run(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "happy path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := _assert.New(t)

			hbc := make(chan datacollector.Heartbeat, 1)
			hbc <- datacollector.Heartbeat{App: "Ableton", Timestamp: time.Now()}
			close(hbc)
			dc := datacollector.New(datacollector.WithHeartbeatChannel(hbc))

			dc.Run()

			buckets := dc.Buckets()
			assert.Equal(1, len(buckets["Ableton"]))
		})
	}
}

func Test_Flush(t *testing.T) {
	t.Setenv("ENV", "test")

	tests := []struct {
		name string
	}{
		{
			name: "happy path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := _assert.New(t)
			sqliteFpath := "../db.test.sqlite"
			d, err := gorm.Open(sqlite.Open(sqliteFpath), &gorm.Config{})
			d.Exec("DELETE FROM sessions;")

			assert.NoError(err)
			sess := gorm.G[models.Session](d)

			basisTime := time.Now().Add(-time.Hour)
			btMinus30 := basisTime.Add(-(30 * time.Minute))
			btMinus29 := basisTime.Add(-(29 * time.Minute))

			dc := datacollector.New(
				datacollector.WithBuckets(map[string][]datacollector.Heartbeat{
					"Ableton": {
						{App: "Ableton", Timestamp: btMinus30},
						{App: "Ableton", Timestamp: btMinus29},
					},
				}),
				datacollector.WithSQLiteFilePath(sqliteFpath),
			)
			dc.Flush()

			abletonSessions, err := sess.Where("application = ?", "Ableton").Find(context.TODO())
			assert.NoError(err)

			assert.Equal(2, len(abletonSessions))
			d.Exec("DELETE FROM sessions;")
		})
	}
}
