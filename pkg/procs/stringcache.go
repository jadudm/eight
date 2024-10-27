package procs

import (
	"time"

	"github.com/patrickmn/go-cache"
)

func StringCache(ch_key <-chan string, ch_value chan<- string, ch_insert chan map[string]string) {
	// Items expire after 60m, purge every 10m.
	c := cache.New(60*time.Minute, 10*time.Minute)

	for {
		select {
		// Ask if things are in the cache
		case key := <-ch_key:
			value, found := c.Get(key)
			if found {
				ch_value <- value.(string)
			} else {
				ch_value <- ""
			}
		// Channel to insert values
		case pairs := <-ch_insert:
			for k, v := range pairs {
				c.Set(k, v, 0)
			}
		}
	}
}
