package collector

import (
	"reflect"
	"testing"
)

func TestEvalAsMetric(t *testing.T) {
	rs := [][]string{
		[]string{"3024", "foo", "1h"},
		[]string{"534", "bar", "1h"},
		[]string{"1521", "foo", "30m"},
		[]string{"231", "bar", "30m"},
	}
	lvs, err := EvalAsMetric(rs...)
	if err != nil {
		t.Error(err)
	}
	expected := []LabeledValue{
		LabeledValue{float64(3024), []string{"foo", "1h"}},
		LabeledValue{float64(534), []string{"bar", "1h"}},
		LabeledValue{float64(1521), []string{"foo", "30m"}},
		LabeledValue{float64(231), []string{"bar", "30m"}},
	}
	if !reflect.DeepEqual(lvs, expected) {
		t.Errorf("actual: %v, expected: %v", lvs, expected)
	}
}
