package collector

import (
	"context"
	"database/sql"
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type sqlCollector struct {
	querier Querier
	metrics SQLMetrics
}

type SQLMetrics []struct {
	Desc      *prometheus.Desc
	SQL       string
	ValueType prometheus.ValueType
	Eval      func(rs ...[]string) ([]LabeledValue, error)
}

type LabeledValue struct {
	Value       float64
	LabelValues []string
}

func EvalAsMetric(rs ...[]string) ([]LabeledValue, error) {
	lvs := make([]LabeledValue, 0)
	for _, r := range rs {
		var v float64
		var vs []string
		if n, err := strconv.ParseFloat(r[0], 64); err != nil {
			return nil, err
		} else {
			v = n
			vs = r[1:]
		}
		lvs = append(lvs, LabeledValue{Value: v, LabelValues: vs})
	}
	return lvs, nil
}

type Querier interface {
	Query(s string) ([][]string, error)
}

type mockQuerier struct {
	f func(s string) ([][]string, error)
}

func NewMockQuerier(f func(s string) ([][]string, error)) *mockQuerier {
	return &mockQuerier{f: f}
}

func (q *mockQuerier) Query(s string) ([][]string, error) {
	return q.f(s)
}

type dbQuerier struct {
	db *sql.DB
}

func NewDBQuerier(db *sql.DB) *dbQuerier {
	return &dbQuerier{db: db}
}

func (q *dbQuerier) Query(s string) ([][]string, error) {
	ctx := context.TODO()
	conn, err := q.db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rs, err := conn.QueryContext(ctx, s)
	if err != nil {
		return nil, err
	}
	defer rs.Close()

	xs := make([][]string, 0)
	numCols := 0
	for rs.Next() {
		if numCols == 0 {
			if cols, err := rs.Columns(); err != nil {
				return nil, err
			} else {
				numCols = len(cols)
			}
		}
		ys := make([]string, numCols)
		ysp := make([]interface{}, len(ys))
		for i, _ := range ys {
			ysp[i] = &ys[i]
		}
		if err := rs.Scan(ysp...); err != nil {
			return nil, err
		}
		xs = append(xs, ys)
	}
	return xs, nil
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
