package main

import (
	"net/http"
	"time"

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
			t := time.Now()
			return [][]string{
				[]string{
					t.Format("150405"),
					time.Local.String(),
				},
				[]string{
					t.UTC().Format("150405"),
					time.UTC.String(),
				},
			}, nil
		},
	)

	mt := collector.SQLMetrics{
		{
			Desc:      prometheus.NewDesc("now_hour_minute", "", []string{"timezone"}, nil),
			SQL:       "SELECT 1",
			ValueType: prometheus.GaugeValue,
			Eval:      collector.EvalAsMetric,
		},
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewSQLCollector(q, mt, 5*time.Second))

	http.Handle("/metrics", metricHandler(registry))
	http.ListenAndServe("localhost:2112", nil)
}
