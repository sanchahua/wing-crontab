package prometheus_stat

import "github.com/prometheus/client_golang/prometheus"

func RegisterGaugeVec(namespace, subSystem, name, help string, metrics []string) *prometheus.GaugeVec {
	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      name,
			Help: help,
		}, metrics)
	prometheus.MustRegister(gauge)
	return gauge
}

func GaugeVecInc(vec *prometheus.GaugeVec, metrics ...string) {
	vec.WithLabelValues(metrics...).Inc()
}

func GaugeVecAdd(vec *prometheus.GaugeVec, value float64, metrics ...string) {
	vec.WithLabelValues(metrics...).Add(value)
}

func GaugeVecSet(vec *prometheus.GaugeVec, value float64, metrics ...string) {
	vec.WithLabelValues(metrics...).Set(value)
}