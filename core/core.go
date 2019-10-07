package core

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type MetricValue struct {
	Value       float64
	LabelValues []string
}

type Store interface {
	Query(ctx context.Context, s string) ([][]string, error)
	Close() error
}

type Recorder interface {
	Record(ctx context.Context) error
}

func EvalAsMetricValues(rs ...[]string) ([]MetricValue, error) {
	mvs := make([]MetricValue, 0)
	for _, r := range rs {
		var v float64
		var vs []string
		if n, err := strconv.ParseFloat(r[0], 64); err != nil {
			return nil, err
		} else {
			v = n
			vs = r[1:]
		}
		mvs = append(mvs, MetricValue{Value: v, LabelValues: vs})
	}
	return mvs, nil
}

func ParseMockRows(s string) ([][]string, error) {
	lines := strings.Split(s, "\n")
	rs := make([][]string, 0)
	nfl := 0
	for _, line := range lines {
		ln := strings.TrimSpace(line)
		if len(ln) == 0 {
			continue
		}
		vs := strings.Split(ln, ",")
		nl := len(vs)
		if nl == 0 {
			continue
		}
		if nfl == 0 {
			nfl = nl
		} else {
			if nfl != nl {
				return nil, errors.New(
					fmt.Sprintf("The number of items doesn't mach the first line. '%s'", line))
			}
		}
		rs = append(rs, vs)
	}
	return rs, nil
}
