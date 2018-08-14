package statistics

import (
	"time"
)

type TimingCollector struct {
	data map[string]*Timing
	defaultWindow int64
}

func NewTimingCollector(defaultWindow int64) *TimingCollector {
	return &TimingCollector{
		data : make(map[string]*Timing),
		defaultWindow: defaultWindow,
	}
}

func (collector *TimingCollector) Add(key string, duration time.Duration) {
	timing := collector.getTiming(key)
	timing.Add(duration)
}

func (collector* TimingCollector) GetAllTimingResult() map[string]*TimingResult{
	result := make(map[string]*TimingResult)
	for key, value := range collector.data {
		result[key] = &TimingResult{Avg:value.Mean()}
	}

	return result
}


func (collector* TimingCollector) getTiming(key string) *Timing {
	if timing, ok := collector.data[key]; ok {
		return timing
	}

	timing := NewTiming(collector.defaultWindow)
	collector.data[key] = timing
	return timing
}