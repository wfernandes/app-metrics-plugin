package parser

type NoOp struct{}

func NewNoOp() *NoOp {
	return &NoOp{}
}

func (n *NoOp) Parse(b []byte) ([]byte, error) {
	return b, nil
}
