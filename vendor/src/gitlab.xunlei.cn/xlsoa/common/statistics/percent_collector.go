package statistics

import (
	"time"
)

type PercentResult struct {
	Percent float64
}

type percentBucket struct {
	numerator *Number
	denominator *Number
}

type PercentCollector struct {
	data map[string]*percentBucket
	defaultWindow      int64
}

func NewPercentCollector(defaultWindow int64) *PercentCollector {
	return &PercentCollector{
		data: make(map[string]*percentBucket),
		defaultWindow: defaultWindow,
	}
}


func (collector *PercentCollector) Inc(key string, isNumerator bool, value float64) {
	bucket := collector.getBucket(key, collector.defaultWindow)
	if isNumerator {
		bucket.numerator.Increment(value)
	} else {
		bucket.denominator.Increment(value)
	}
}

func (collector* PercentCollector) GetAllPercentResult() map[string]*PercentResult{
	result := make(map[string]*PercentResult)
	now := time.Now()
	for key, value := range collector.data {
		numeratorResult := value.numerator.CalcNumberResult(now)
		denominatorResult := value.denominator.CalcNumberResult(now)
		result[key] = &PercentResult{
			Percent: collector.calcPercent(numeratorResult.Sum, denominatorResult.Sum),
		}
	}
	return result
}

func (collector *PercentCollector) calcPercent(numerator, denominator float64) float64 {
	if denominator == 0.0 {
		return 0.0
	}
	return numerator/denominator
}

func (collector* PercentCollector) getBucket(key string, window int64) *percentBucket{
	if bucket, ok := collector.data[key]; ok {
		return bucket
	}

	bucket := &percentBucket{
		numerator :  NewNumber(window),
		denominator:  NewNumber(window),
	}
	collector.data[key] = bucket
	return bucket
}
