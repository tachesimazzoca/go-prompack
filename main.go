package main

import (
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tachesimazzoca/go-prompack/prompack"
)

func metricHandler(registry *prometheus.Registry) http.Handler {
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
}

func main() {
	q := prompack.NewMockQuerier(
		func(s string) ([][]string, error) {
			return [][]string{
				[]string{
					strconv.Itoa(rand.Intn(500) + rand.Intn(500)),
					"foo",
				},
				[]string{
					strconv.Itoa(rand.Intn(500) + rand.Intn(500)),
					"bar",
				},
				[]string{
					strconv.Itoa(rand.Intn(500) + rand.Intn(500)),
					"baz",
				},
			}, nil
		},
	)

	registry := prometheus.NewRegistry()
	labelNames := []string{"job"}
	gvec := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "process_time",
	}, labelNames)
	registry.MustRegister(gvec)

	hvec := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "process_time_histogram",
		Buckets: []float64{150, 850},
	}, labelNames)
	registry.MustRegister(hvec)

	svec := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "process_time_summary",
	}, labelNames)
	registry.MustRegister(svec)

	m := prompack.NewSQLMeasurer(
		q, "SELECT 1", func(lv prompack.LabeledValue) {
			gvec.WithLabelValues(lv.LabelValues...).Set(lv.Value)
			hvec.WithLabelValues(lv.LabelValues...).Observe(lv.Value)
			svec.WithLabelValues(lv.LabelValues...).Observe(lv.Value)
		})

	go func() {
		for {
			m.Measure()
			time.Sleep(5 * time.Second)
		}
	}()

	http.Handle("/metrics", metricHandler(registry))
	http.ListenAndServe("localhost:2112", nil)
}
