package statistics

import (
	"time"
)

type NumberCollector struct {
	data map[string]*Number
	defaultWindow      int64
}

func NewNumberCollector(defaultWindow int64) *NumberCollector {
	return &NumberCollector{
		data: make(map[string]*Number),
		defaultWindow: defaultWindow,
	}
}

func (collector *NumberCollector) Inc(key string, value float64) {
	number := collector.getNumber(key, collector.defaultWindow)
	number.Increment(value)
}

func (collector* NumberCollector) GetAllNumberResult() map[string]*NumberResult{
	result := make(map[string]*NumberResult)
	now := time.Now()
	for key, value := range collector.data {
		result[key] = value.CalcNumberResult(now)
	}
	return result
}


func (collector* NumberCollector) getNumber(key string, window int64) *Number {
	if number, ok := collector.data[key]; ok {
		return number
	}

	number := NewNumber(window)
	collector.data[key] = number
	return number
}
