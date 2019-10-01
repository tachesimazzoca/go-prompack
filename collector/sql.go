package collector

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type sqlCollector struct {
	mux      sync.Mutex
	querier  Querier
	metrics  SQLMetrics
	mtCache  [][]LabeledValue
	interval time.Duration
}

func NewSQLCollector(q Querier, mt SQLMetrics, d time.Duration) prometheus.Collector {
	c := &sqlCollector{
		querier:  q,
		metrics:  mt,
		mtCache:  make([][]LabeledValue, len(mt)),
		interval: d,
	}
	c.run()
	return c
}

func (c *sqlCollector) run() {
	go func() {
		for c.interval > 0 {
			log.Println("Collecting ...")
			c.updateCache()
			time.Sleep(c.interval)
		}
	}()
}

func (c *sqlCollector) updateCache() {
	for i, mt := range c.metrics {
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
		c.mtCache[i] = lvs
	}
}

func (c *sqlCollector) Collect(ch chan<- prometheus.Metric) {
	defer c.mux.Unlock()
	if c.interval == 0 {
		c.updateCache()
	}
	c.mux.Lock()
	for i, lvs := range c.mtCache {
		mt := c.metrics[i]
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
