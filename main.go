package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tachesimazzoca/go-prompack/collector"
)

func metricHandler(registry *prometheus.Registry) http.Handler {
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
}

func main() {
	q := collector.NewMockQuerier(
		func(s string) ([][]string, error) {
			return [][]string{}, nil
		},
	)

	mt := collector.SQLMetrics{}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewSQLCollector(q, mt))

	http.Handle("/metrics", metricHandler(registry))
	http.ListenAndServe("localhost:2112", nil)
}
