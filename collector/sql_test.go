package collector

import (
	"reflect"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/prometheus/client_golang/prometheus"
)

func TestSQLCollector(t *testing.T) {
	q := NewMockQuerier(
		func(s string) ([][]string, error) {
			return [][]string{
				[]string{"3024", "foo", "1h"},
				[]string{"534", "bar", "1h"},
			}, nil
		},
	)
	mt := SQLMetrics{
		{
			Desc:      prometheus.NewDesc("num_orders", "The number of placed orders in the last hour.", []string{"site", "interval"}, nil),
			SQL:       "SELECT COUNT(1), site, '1h' FROM purchase_orders WHERE created_at < NOW() - INTERVAL 1 HOUR GROUP BY site",
			ValueType: prometheus.GaugeValue,
			Eval:      EvalAsMetric,
		},
	}
	c := NewSQLCollector(q, mt)
	ch := make(chan prometheus.Metric)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	expected := []LabeledValue{
		LabeledValue{Value: 3024, LabelValues: []string{"1h", "foo"}},
		LabeledValue{Value: 534, LabelValues: []string{"1h", "bar"}},
	}

	for _, x := range expected {
		mt := <-ch
		pb := &dto.Metric{}
		mt.Write(pb)
		g := pb.GetGauge()
		if v := g.GetValue(); v != x.Value {
			t.Errorf("actual: %f, expected: %f", v, x.Value)
		}
		lvs := make([]string, 0)
		for _, lp := range pb.GetLabel() {
			lvs = append(lvs, lp.GetValue())
		}
		if !reflect.DeepEqual(lvs, x.LabelValues) {
			t.Errorf("actual: %v, expected: %v", lvs, x.LabelValues)
		}
	}
}
