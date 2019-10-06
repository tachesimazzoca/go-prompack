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
	querierMap := map[string]prompack.Querier{
		"default": q,
	}
	labelNames := []string{"job"}
	collectorMap := map[string]prometheus.Collector{
		"process_time": prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "process_time",
		}, labelNames),
		"process_time_histogram": prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "process_time_histogram",
			Buckets: []float64{150, 850},
		}, labelNames),
		"process_time_summary": prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name: "process_time_summary",
		}, labelNames),
	}
	measurerOptsMap := map[string]prompack.MeasurerOpts{
		"process_time": prompack.MeasurerOpts{
			Name:        "process_time",
			QuerierName: "default",
			QueryString: "SELECT 1",
			MetricNames: []string{"process_time", "process_time_histogram", "process_time_summary"},
			Interval:    5 * time.Second,
		},
	}

	ws := prompack.NewWorkspace(querierMap, collectorMap, measurerOptsMap)
	registry := prometheus.NewRegistry()
	ws.Start(registry)

	http.Handle("/metrics", metricHandler(registry))
	http.ListenAndServe("localhost:2112", nil)
}
