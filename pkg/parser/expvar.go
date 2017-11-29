package parser

import "encoding/json"

type Expvar struct {
	propsToRemove []string
}

type ExpvarOpt func(*Expvar)

func WithPropertiesToRemove(toRemove []string) ExpvarOpt {
	return func(e *Expvar) {
		e.propsToRemove = toRemove
	}
}

func NewExpvar(opts ...ExpvarOpt) *Expvar {
	e := &Expvar{}

	for _, o := range opts {
		o(e)
	}

	return e
}

func (e *Expvar) Parse(b []byte) (map[string]interface{}, error) {
	output := make(map[string]interface{})
	err := json.Unmarshal(b, &output)
	if err != nil {
		return nil, err
	}

	for _, prop := range e.propsToRemove {
		delete(output, prop)
	}

	return output, nil
}
