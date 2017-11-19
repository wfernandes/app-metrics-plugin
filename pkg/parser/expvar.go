package parser

import "encoding/json"

type Expvar struct {
	propsToRemove []string
}

func NewExpvar(toRemove []string) *Expvar {
	return &Expvar{
		propsToRemove: toRemove,
	}
}

func (e *Expvar) Parse(b []byte) ([]byte, error) {
	output := make(map[string]*json.RawMessage)
	err := json.Unmarshal(b, &output)
	if err != nil {
		return nil, err
	}

	for _, prop := range e.propsToRemove {
		delete(output, prop)
	}

	return json.Marshal(output)
}
