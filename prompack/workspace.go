package prompack

import (
	"errors"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MeasurerOpts struct {
	Name        string
	QuerierName string
	QueryString string
	MetricNames []string
	Interval    time.Duration
}

type Workspace struct {
	querierMap      map[string]Querier
	collectorMap    map[string]prometheus.Collector
	measurerOptsMap map[string]MeasurerOpts

	mux               sync.Mutex
	started           bool
	measurerTickerMap map[string]*time.Ticker
	measurerDoneMap   map[string]chan bool
}

func NewWorkspace(qm map[string]Querier, cm map[string]prometheus.Collector,
	mom map[string]MeasurerOpts) *Workspace {
	return &Workspace{
		started:         false,
		querierMap:      qm,
		collectorMap:    cm,
		measurerOptsMap: mom,
	}
}

func (ws *Workspace) Start(r prometheus.Registerer) error {
	defer ws.mux.Unlock()
	ws.mux.Lock()
	if ws.started {
		return errors.New("The workspace has already started")
	}
	ws.started = true
	for _, v := range ws.collectorMap {
		r.MustRegister(v)
	}
	ws.measurerTickerMap = make(map[string]*time.Ticker)
	ws.measurerDoneMap = make(map[string]chan bool)
	for k, v := range ws.measurerOptsMap {
		ws.measurerTickerMap[k] = time.NewTicker(v.Interval)
		if _, ok := ws.measurerDoneMap[k]; !ok {
			ws.measurerDoneMap[k] = make(chan bool)
		}
		m := NewSQLMeasurer(ws.querierMap[v.QuerierName], v.QueryString, func(lv LabeledValue) {
			for _, n := range v.MetricNames {
				switch c := ws.collectorMap[n].(type) {
				case *prometheus.GaugeVec:
					c.WithLabelValues(lv.LabelValues...).Set(lv.Value)
				case *prometheus.HistogramVec:
					c.WithLabelValues(lv.LabelValues...).Observe(lv.Value)
				case *prometheus.SummaryVec:
					c.WithLabelValues(lv.LabelValues...).Observe(lv.Value)
				default:
				}
			}
		})
		go func(done <-chan bool, tick <-chan time.Time) {
			for {
				select {
				case <-done:
					return
				case <-tick:
					m.Measure()
				}
			}
		}(ws.measurerDoneMap[k], ws.measurerTickerMap[k].C)
	}
	return nil
}

func (ws *Workspace) Stop() error {
	defer ws.mux.Unlock()
	ws.mux.Lock()
	if !ws.started {
		return errors.New("The workspace has not started yet.")
	}
	for _, v := range ws.measurerTickerMap {
		v.Stop()
	}
	for _, ch := range ws.measurerDoneMap {
		ch <- true
	}
	return nil
}
