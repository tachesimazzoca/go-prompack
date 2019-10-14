package core

import (
	"time"

	yaml "gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Addr string
}

type StoreConfig struct {
	Type   string
	Params map[string]string
}

type MetricConfig struct {
	Type       string
	Name       string
	Help       string
	Buckets    []float64
	Objectives map[float64]float64
	LabelNames []string `yaml:"labelNames"`
}

type RecorderConfig struct {
	Type        string
	StoreName   string `yaml:"storeName"`
	Query       string
	Interval    time.Duration
	MetricNames []string `yaml:"metricNames"`
}

type ExporterConfig struct {
	Server    ServerConfig
	Store     map[string]StoreConfig
	Metrics   []MetricConfig
	Recorders []RecorderConfig
}

func NewExporterConfigFromYaml(b []byte) (ExporterConfig, error) {
	c := ExporterConfig{}
	if err := yaml.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}
