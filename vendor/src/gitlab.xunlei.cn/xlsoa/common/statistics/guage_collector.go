package statistics

type gaugeBucket struct {
	value float64
}

func (bucket *gaugeBucket) add(value float64) {
	bucket.value += value
}

func (bucket *gaugeBucket) getAndReset() float64 {
	value := bucket.value
	bucket.value = 0.0
	return value
}

type GaugeCollector struct {
	data map[string]*gaugeBucket
}

func NewGaugeCollector() *GaugeCollector {
	return &GaugeCollector{
		data: make(map[string]*gaugeBucket),
	}
}

func (collector *GaugeCollector) Inc(key string, value float64) {
	bucket := collector.getBucket(key)
	bucket.value += value
}

func (collector* GaugeCollector) GetAllGauge() map[string]float64{
	result := make(map[string]float64)
	for key, value := range collector.data {
		result[key] = value.getAndReset()
	}
	return result
}

func (collector* GaugeCollector) getBucket(key string) *gaugeBucket {
	if bucket, ok := collector.data[key]; ok {
		return bucket
	}

	bucket := &gaugeBucket{}
	collector.data[key] = bucket
	return bucket
}
