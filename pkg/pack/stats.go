package pack

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

type PackStats struct {
	BaseStats
}

var ES *PackStats

func NewPackStats() {
	if ES == nil {
		e := PackStats{}
		e.stats = sync.Map{}
		ES = &e
	}
}

// pack     | fatal error: concurrent map writes
func (e *PackStats) Set(key string, val int64) {
	e.stats.Store(key, val)
}

func (e *PackStats) Get(key string) int64 {
	v, _ := e.stats.Load(key)
	return v.(int64)
}

func (e *PackStats) GetAll() map[string]int64 {
	// func (m *Map) Range(f func(key, value any) bool)
	copy := make(map[string]int64, 0)
	e.stats.Range(func(key any, v any) bool {
		copy[key.(string)] = v.(int64)
		return true
	})
	return copy
}

func (e *PackStats) Increment(key string) {
	if val, ok := e.stats.Load(key); ok {
		e.Set(key, val.(int64)+1)
	} else {
		e.Set(key, 1)
	}
}
