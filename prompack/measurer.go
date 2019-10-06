package prompack

type Measurer interface {
	Measure() error
}

type sqlMeasurer struct {
	querier     Querier
	queryString string
	measureF    func(lv LabeledValue)
}

func NewSQLMeasurer(q Querier, qs string, f func(lv LabeledValue)) *sqlMeasurer {
	return &sqlMeasurer{
		querier:     q,
		queryString: qs,
		measureF:    f,
	}
}

func (m *sqlMeasurer) Measure() error {
	rs, err := m.querier.Query(m.queryString)
	if err != nil {
		return err
	}
	lvs, err := evalAsLabeledValues(rs...)
	if err != nil {
		return err
	}
	for _, lv := range lvs {
		m.measureF(lv)
	}
	return nil
}
