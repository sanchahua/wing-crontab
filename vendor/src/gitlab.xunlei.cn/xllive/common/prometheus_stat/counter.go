package prometheus_stat

import "github.com/prometheus/client_golang/prometheus"

func RegisterCounterVec(namespace, subSystem, name, help string, metrics []string) *prometheus.CounterVec {
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      name,
			Help: help,
		}, metrics)
	prometheus.MustRegister(counter)
	return counter
}

func CounterVecInc(vec *prometheus.CounterVec, metrics ...string) {
	vec.WithLabelValues(metrics...).Inc()
}

func CounterVecAdd(vec *prometheus.CounterVec, value float64, metrics ...string) {
	vec.WithLabelValues(metrics...).Add(value)
}