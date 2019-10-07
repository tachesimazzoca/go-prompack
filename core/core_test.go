package core

import (
	"reflect"
	"testing"
)

func TestEvalAsMetricValues(t *testing.T) {
	rs := [][]string{
		[]string{"3024", "foo", "1h"},
		[]string{"534", "bar", "1h"},
		[]string{"1521", "foo", "30m"},
		[]string{"231", "bar", "30m"},
	}
	mvs, err := EvalAsMetricValues(rs...)
	if err != nil {
		t.Error(err)
	}
	expected := []MetricValue{
		MetricValue{float64(3024), []string{"foo", "1h"}},
		MetricValue{float64(534), []string{"bar", "1h"}},
		MetricValue{float64(1521), []string{"foo", "30m"}},
		MetricValue{float64(231), []string{"bar", "30m"}},
	}
	if !reflect.DeepEqual(mvs, expected) {
		t.Errorf("actual: %v, expected: %v", mvs, expected)
	}
}
