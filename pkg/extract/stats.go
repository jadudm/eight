package extract

import (
	"sync"
)

// FIXME Switch to a concurrency-safe map library...
type StatsMap = sync.Map

type Stats interface {
	Set(string, int64)
	Increment(string)
	Get(string) int64
	GetAll() StatsMap
}

type BaseStats struct {
	stats StatsMap
}

type ExtractStats struct {
	BaseStats
}

var ES *ExtractStats

func NewExtractStats() {
	if ES == nil {
		e := ExtractStats{}
		e.stats = sync.Map{}
		ES = &e
	}
}

// extract     | fatal error: concurrent map writes
func (e *ExtractStats) Set(key string, val int64) {
	e.stats.Store(key, val)
}

func (e *ExtractStats) Get(key string) int64 {
	v, _ := e.stats.Load(key)
	return v.(int64)
}

func (e *ExtractStats) GetAll() map[string]int64 {
	// func (m *Map) Range(f func(key, value any) bool)
	copy := make(map[string]int64, 0)
	e.stats.Range(func(key any, v any) bool {
		copy[key.(string)] = v.(int64)
		return true
	})
	return copy
}

func (e *ExtractStats) Increment(key string) {
	if val, ok := e.stats.Load(key); ok {
		e.Set(key, val.(int64)+1)
	} else {
		e.Set(key, 1)
	}
}
