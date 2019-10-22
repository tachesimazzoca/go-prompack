package core

import (
	"time"

	yaml "gopkg.in/yaml.v2"
)

// ServerConfig represents the "server" section of config.yml. It has some
// parameters corresponding to http.Server.
//
// An example configuration would be like the following:
//
//   server:
//     addr: "localhost:2112"
//
type ServerConfig struct {
	Addr string
}

// StoreConfig represents the map values of "store" section of config.yml.  
// Those parameters will be used to create instances of core.Store.
//
// An example configuration would be like the following:
//
//   store:
//     default:
//       type: db
//       params:
//         driverName: mysql
//         dataSourceName: "username:password@tcp(127.0.0.1:3306)/database"
//     mock:
//       type: mock
//       params:
//         rows: |
//           1,foo
//           2,bar
//
type StoreConfig struct {
	Type   string
	Params map[string]string
}

// MetricConfig represents an element of the "metrics" section of config.yml.
// Those parameters will be used to create instances of prometheus.Collector.
//
// An example configuration would be like the following:
//
//   metrics:
//     - type: gauge
//       name: num_transactions
//       help: Number of transactions
//       labelNames: ["interval"]
//
type MetricConfig struct {
	Type       string
	Name       string
	Help       string
	Buckets    []float64
	Objectives map[float64]float64
	LabelNames []string `yaml:"labelNames"`
}

// RecorderConfig represents an element of the "recorders" section of config.yml.
// Those parameters will be used to create instances of core.Recorder.
//
// An example configuration would be like the following:
//
//   recorders:
//     - type: sql
//       storName: default
//       query: |
//           SELECT COUNT(1), '1d' FROM transactions
//           WHERE created_at >= NOW() - INTERVAL 1 HOUR
//       interval: 1m
//       metricNames: ["num_transactions"]
//
type RecorderConfig struct {
	Type        string
	StoreName   string `yaml:"storeName"`
	Query       string
	Interval    time.Duration
	MetricNames []string `yaml:"metricNames"`
}

// ExporterConfig represents a whole parts of config.yml as settings given
// to the utility function exporter.NewExporter.
type ExporterConfig struct {
	Server    ServerConfig
	Store     map[string]StoreConfig
	Metrics   []MetricConfig
	Recorders []RecorderConfig
}

// NewExporterConfigFromYaml converts config.yml as an array of byte into
// an instance of core.ExporterConfig. 
func NewExporterConfigFromYaml(b []byte) (ExporterConfig, error) {
	c := ExporterConfig{}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}
