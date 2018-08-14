package statistics


type counterBucket struct {
	Value float64
}

type CounterCollector struct {
	data map[string]*counterBucket
}

func NewCounterCollector() *CounterCollector {
	return &CounterCollector{
		data: make(map[string]*counterBucket),
	}
}


func (collector *CounterCollector) Inc(key string, value float64) {
	bucket := collector.getBucket(key)
	bucket.Value += value
}

func (collector* CounterCollector) GetAllCounter() map[string]float64{
	result := make(map[string]float64)
	for key, value := range collector.data {
		result[key] = value.Value
	}
	return result
}


func (collector* CounterCollector) getBucket(key string) *counterBucket {
	if bucket, ok := collector.data[key]; ok {
		return bucket
	}
	bucket := &counterBucket{}
	collector.data[key] = bucket
	return bucket
}
