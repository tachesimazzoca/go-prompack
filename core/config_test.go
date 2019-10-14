package core

import (
	"reflect"
	"testing"
	"time"
)

func TestNewExporterConfigFromYaml(t *testing.T) {
	yml := `
server:
  addr: "localhost:2112"

store:
  default:
    type: db
    params:
      driverName: mysql
      dataSourceName: "test:test@localhost/test"
metrics:
  - name: foo
    help: Foo is the first metric.
    type: gauge
    labelNames: 
      - num_threads
  - name: bar
    help: Bar is the second metric.
    type: histogram
    buckets: [0.1, 3.25, 5, 8.5, 10]
    labelNames: [num_processes]
  - name: baz
    help: Baz is the third metric.
    type: summary
    objectives:
      0.5: 0.05
      0.9: 0.01
      0.99: 0.001
recorders:
  - type: sql
    storeName: default
    interval: 5s
    metricNames: [foo, bar]
    query: "SELECT NOW() FROM dual WHERE 1 = 1"
`
	actual, err := NewExporterConfigFromYaml([]byte(yml))
	if err != nil {
		t.Error(err)
	}
	expected := ExporterConfig{
		Server: ServerConfig{
			Addr: "localhost:2112",
		},
		Store: map[string]StoreConfig{
			"default": StoreConfig{
				Type: "db",
				Params: map[string]string{
					"driverName":     "mysql",
					"dataSourceName": "test:test@localhost/test",
				},
			},
		},
		Metrics: []MetricConfig{
			MetricConfig{
				Type:       "gauge",
				Name:       "foo",
				Help:       "Foo is the first metric.",
				LabelNames: []string{"num_threads"},
			},
			MetricConfig{
				Type:       "histogram",
				Name:       "bar",
				Help:       "Bar is the second metric.",
				Buckets:    []float64{0.1, 3.25, 5, 8.5, 10},
				LabelNames: []string{"num_processes"},
			},
			MetricConfig{
				Type: "summary",
				Name: "baz",
				Help: "Baz is the third metric.",
				Objectives: map[float64]float64{
					0.5:  0.05,
					0.9:  0.01,
					0.99: 0.001,
				},
			},
		},
		Recorders: []RecorderConfig{
			RecorderConfig{
				Type:        "sql",
				StoreName:   "default",
				Query:       "SELECT NOW() FROM dual WHERE 1 = 1",
				Interval:    5 * time.Second,
				MetricNames: []string{"foo", "bar"},
			},
		},
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("actual: %v, expected: %v", actual, expected)
	}
}
