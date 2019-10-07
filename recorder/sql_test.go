package recorder

import (
	"context"
	"reflect"
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/tachesimazzoca/go-prompack/core"
	"github.com/tachesimazzoca/go-prompack/store"

	"github.com/prometheus/client_golang/prometheus"
)

func TestSQLRecorder(t *testing.T) {
	st := store.NewMockStore(
		func(s string) ([][]string, error) {
			return [][]string{
				[]string{"3024", "foo", "1h"},
			}, nil
		},
	)
	c := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "num_orders",
		Help: "The number of placed orders in the last hour.",
	}, []string{"site", "interval"})
	rec := NewSQLRecorder(
		st, "SELECT COUNT(1), site, '1h' FROM purchase_orders"+
			" WHERE created_at < NOW() - INTERVAL 1 HOUR GROUP BY site",
		func(mv core.MetricValue) {
			c.WithLabelValues(mv.LabelValues...).Set(mv.Value)
		})
	rec.Record(context.TODO())
	ch := make(chan prometheus.Metric)
	go func() {
		c.Collect(ch)
		close(ch)
	}()

	expected := []core.MetricValue{
		core.MetricValue{Value: 3024, LabelValues: []string{"1h", "foo"}},
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
