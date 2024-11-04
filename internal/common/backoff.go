package common

import (
	"math/rand/v2"
	"sync"
	"time"

	"go.uber.org/zap"
)

func BackoffLoop(host string, polite_sleep_milliseconds int64, last_hit *sync.Map, last_backoff *sync.Map) {
	for {
		// Look at the timeing map.
		last_hit_time, ok := last_hit.Load(host)
		// If we're in the map, and we're within 2s, we should keep checking after a backoff
		polite_duration := time.Duration(polite_sleep_milliseconds) * time.Millisecond

		if ok && (time.Since(last_hit_time.(time.Time)) < polite_duration) {
			// There will be a last backoff time.
			last, _ := last_backoff.Load(host)
			new_backoff_time := int64(float64(polite_sleep_milliseconds)/10*rand.Float64()) + int64(float64(last.(int64))*1.03)
			// Go back to sleep
			zap.L().Debug("backing off and sleeping", zap.String("host", host), zap.Int64("duration", new_backoff_time))
			time.Sleep(time.Duration(new_backoff_time) * time.Millisecond)
			continue
		} else {
			// We're not in the map, or it is more than <polite> milliseconds!
			// IT IS OUR TURN.
			// Reset the times and get out of here.
			zap.L().Debug("Freedom '90")
			last_backoff.Store(host, polite_sleep_milliseconds)
			last_hit.Store(host, time.Now())
			break
		}
	}

}
