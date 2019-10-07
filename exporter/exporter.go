package exporter

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tachesimazzoca/go-prompack/core"
	"github.com/tachesimazzoca/go-prompack/recorder"
	"github.com/tachesimazzoca/go-prompack/store"
)

type Exporter struct {
	storeMap     map[string]core.Store
	collectorMap map[string]prometheus.Collector
	recorders    []core.Recorder

	mux     sync.Mutex
	started bool
	tickers []*time.Ticker
	quitCs  []chan struct{}
}

func NewExporter(cfg core.ExporterConfig) (*Exporter, error) {

	log.Printf("Loading configuration: %#v", cfg)

	sm := make(map[string]core.Store, len(cfg.Store))
	for k, v := range cfg.Store {
		switch v.Type {
		case "mock":
			if s, ok := v.Params["rows"]; ok {
				if rows, err := core.ParseMockRows(s); err != nil {
					return nil, err
				} else {
					sm[k] = store.NewMockStore(func(s string) ([][]string, error) {
						return rows, nil
					})
				}
			}
		case "db":
			db, err := sql.Open(v.Params["driverName"], v.Params["dataSourceName"])
			if err != nil {
				return nil, err
			}
			sm[k] = store.NewDBStore(db)
		default:
		}
	}

	cm := make(map[string]prometheus.Collector, len(cfg.Metrics))
	for _, v := range cfg.Metrics {
		switch v.Type {
		case "counter":
			cm[v.Name] = prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: v.Name,
				Help: v.Help,
			}, v.LabelNames)
		case "gauge":
			cm[v.Name] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: v.Name,
				Help: v.Help,
			}, v.LabelNames)
		case "histogram":
			cm[v.Name] = prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name: v.Name,
				Help: v.Help,
			}, v.LabelNames)
		case "summary":
			cm[v.Name] = prometheus.NewSummaryVec(prometheus.SummaryOpts{
				Name: v.Name,
				Help: v.Help,
			}, v.LabelNames)
		default:
			return nil, errors.New("metrics.type must be in (counter|gauge|histogram|summary)")
		}
	}

	recs := make([]core.Recorder, len(cfg.Recorders))
	ts := make([]*time.Ticker, len(cfg.Recorders))
	qs := make([]chan struct{}, len(cfg.Recorders))
	for i, v := range cfg.Recorders {
		if len(v.MetricNames) == 0 {
			return nil, errors.New(fmt.Sprintf("recorders[%d] has empty metric names", i))
		}
		if st, ok := sm[v.StoreName]; !ok {
			return nil, errors.New(fmt.Sprintf("store %s is missing", v.StoreName))
		} else {
			f := func(mv core.MetricValue) {
				for _, n := range v.MetricNames {
					switch c := cm[n].(type) {
					case *prometheus.CounterVec:
						c.WithLabelValues(mv.LabelValues...).Add(mv.Value)
					case *prometheus.GaugeVec:
						c.WithLabelValues(mv.LabelValues...).Set(mv.Value)
					case *prometheus.HistogramVec:
						c.WithLabelValues(mv.LabelValues...).Observe(mv.Value)
					case *prometheus.SummaryVec:
						c.WithLabelValues(mv.LabelValues...).Observe(mv.Value)
					default:
					}
				}
			}
			switch v.Type {
			case "sql":
				recs[i] = recorder.NewSQLRecorder(st, v.Query, f)
			default:
				return nil, errors.New("recorders[*].type must be in (sql)")
			}
		}
		ts[i] = time.NewTicker(v.Interval)
		qs[i] = make(chan struct{})
	}

	return &Exporter{
		storeMap:     sm,
		collectorMap: cm,
		recorders:    recs,

		started: false,
		tickers: ts,
		quitCs:  qs,
	}, nil
}

func (ep *Exporter) Start(ctx context.Context, r prometheus.Registerer) error {
	log.Printf("Starting exporter %p", ep)
	defer ep.mux.Unlock()
	ep.mux.Lock()
	if ep.started {
		return errors.New("The exporter has already started.")
	}
	ep.started = true
	for _, v := range ep.collectorMap {
		log.Printf("Register %#v", v)
		r.MustRegister(v)
	}
	for i, v := range ep.recorders {
		log.Printf("Start recording with %p", v)
		go func(r core.Recorder, qc <-chan struct{}, tc <-chan time.Time) {
			rec := func() {
				if err := r.Record(ctx); err != nil {
					log.Printf("(%p).Record returned error: %v", r, err)
				}
			}
			rec()
			recording := true
			for recording {
				select {
				case <-qc:
					log.Printf("Recorder stopped: %p", r)
					recording = false
					break
				case <-tc:
					rec()
				}
			}
			log.Printf("Stop recording with %p", r)
		}(v, ep.quitCs[i], ep.tickers[i].C)
	}
	return nil
}

func (ep *Exporter) Stop() error {
	log.Printf("Stopping exporter %p", ep)
	defer ep.mux.Unlock()
	ep.mux.Lock()
	if !ep.started {
		return errors.New("The exporter has not started yet.")
	}
	for _, v := range ep.tickers {
		v.Stop()
	}
	for _, ch := range ep.quitCs {
		close(ch)
	}
	for k, v := range ep.storeMap {
		v.Close()
		log.Printf("Store closed: %s", k)
	}
	return nil
}
