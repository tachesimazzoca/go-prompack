package prompack

import (
	"strconv"
)

type LabeledValue struct {
	Value       float64
	LabelValues []string
}

func evalAsLabeledValues(rs ...[]string) ([]LabeledValue, error) {
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
