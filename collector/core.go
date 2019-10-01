package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type LabeledValue struct {
	Value       float64
	LabelValues []string
}

type SQLMetrics []struct {
	Desc      *prometheus.Desc
	SQL       string
	ValueType prometheus.ValueType
	Eval      func(rs ...[]string) ([]LabeledValue, error)
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
