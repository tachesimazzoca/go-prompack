package collector

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type sqlCollector struct {
	querier Querier
	metrics SQLMetrics
}

func NewSQLCollector(q Querier, mt SQLMetrics) prometheus.Collector {
	return &sqlCollector{
		querier: q,
		metrics: mt,
	}
}

func (c *sqlCollector) Collect(ch chan<- prometheus.Metric) {
	for _, mt := range c.metrics {
		rs, err := c.querier.Query(mt.SQL)
		if err != nil {
			log.Println(err)
			continue
		}
		lvs, err := mt.Eval(rs...)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, lv := range lvs {
			ch <- prometheus.MustNewConstMetric(
				mt.Desc, mt.ValueType, lv.Value, lv.LabelValues...)
		}
	}
}

func (c *sqlCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, mt := range c.metrics {
		ch <- mt.Desc
	}
}
