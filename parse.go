package effe

import "io"

type node struct {
	typ       nodeType
	value     string
	rangeSpec rangeSpec
	children  []node
}

type nodeType int

const (
	Function nodeType = iota
	Range
	Literal
	OperatorPrefix
	OperatorInfix
	OperatorPostfix
)

type parseError struct {
	location int
	message  string
}

func (pe *parseError) Error() string {
	return pe.message
}

func Parse(io.RuneReader) (node, []parseError, error) {
	return node{
		typ:   Function,
		value: "sum",
		children: []node{
			node{
				typ:       Range,
				rangeSpec: makeRowsRangeSpec(1, 1),
			},
		},
	}, nil, nil
}
