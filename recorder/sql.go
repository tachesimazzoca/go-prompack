package recorder

import (
	"context"

	"github.com/tachesimazzoca/go-prompack/core"
)

type sqlRecorder struct {
	store   core.Store
	query   string
	recordF func(mv core.MetricValue)
}

func NewSQLRecorder(s core.Store, q string, f func(mv core.MetricValue)) *sqlRecorder {
	return &sqlRecorder{
		store:   s,
		query:   q,
		recordF: f,
	}
}

func (rec *sqlRecorder) Record(ctx context.Context) error {
	rs, err := rec.store.Query(ctx, rec.query)
	if err != nil {
		return err
	}
	mvs, err := core.EvalAsMetricValues(rs...)
	if err != nil {
		return err
	}
	for _, v := range mvs {
		rec.recordF(v)
	}
	return nil
}
