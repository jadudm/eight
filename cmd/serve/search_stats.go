package main

import (
	"strings"
	"time"

	"github.com/jadudm/eight/internal/common"
)

func runStats(sri ServeRequestInput, duration time.Duration) {

	// Search accounting
	totalStats := common.NewBaseStats("serve")
	totalStats.Increment("queries")
	totalStats.Sum("total_query_time", duration.Nanoseconds())
	if totalStats.HasKey("total_query_time") && totalStats.HasKey("queries") {
		totalStats.Set("average_query_time", int64(totalStats.Get("total_query_time")/totalStats.Get("queries")))
	}

	var stats *common.BaseStats
	if m, ok := statmap.Load(sri.Host); ok {
		stats = m.(*common.BaseStats)
	} else {
		stats = common.NewBaseStats(sri.Host)
		statmap.Store(sri.Host, stats)
	}

	stats.Increment("queries")
	// stats.Increment("_" + sri.Host)
	stats.Sum("total_query_time", duration.Nanoseconds())
	if stats.HasKey("total_query_time") && stats.HasKey("queries") {
		stats.Set("average_query_time", int64(stats.Get("total_query_time")/stats.Get("queries")))
	}

	// Count all the search terms? Why not!
	for _, t := range strings.Split(sri.Terms, " ") {
		stats.Increment("term:" + t)
	}

}
