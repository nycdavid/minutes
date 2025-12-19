package datacollector_test

import (
	"testing"
	"time"

	"github.com/nycdavid/minutes/datacollector"
	_assert "github.com/stretchr/testify/assert"
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
