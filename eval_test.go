package effe

import (
	"strings"
	"testing"
)

// Stubbed range implements a 10x10 range where each cell contains
// a number determined by position: column# * 10 + row# * 1
// type stubbedRange struct{}

// var tenByTenRange = makeRangeSpec(1, 10, 1, 10)

// func (s stubbedRange) iterate(r rangeSpec, closure func(v value) error) error {
// 	concreteRange := tenByTenRange.intersect(r)
// 	for c := concreteRange.cLo; c <= concreteRange.cHi; c++ {
// 		for r := concreteRange.rLo; r <= concreteRange.rHi; r++ {
// 			v := apd.New(int64(10*c+r), 0)
// 			err := closure(numberValue{v: v})
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

type tokenizeTestCase struct {
	name     string
	cell     string
	validate func(t *testing.T, ts []token)
}

var tokenCases []tokenizeTestCase = []tokenizeTestCase{
	tokenizeTestCase{
		name: "sum",
		cell: "=sum(A1:A10)",
		validate: func(t *testing.T, ts []token) {
			if len(ts) != 4 {
				t.Errorf("Expected length 4, but got: %v", ts)
			}
		},
	},
	tokenizeTestCase{
		name: "Reference",
		cell: "=A1",
		validate: func(t *testing.T, ts []token) {
			if len(ts) != 1 {
				t.Errorf("Expected length 1, but got: %v", ts)
			}
			if ts[0] != (token{value: "A1", typ: TokenTypeRange}) {
				t.Errorf("Expected a range, but got %v", ts[0])

			}
		},
	},
	tokenizeTestCase{
		name: "sum column",
		cell: "=sum(A:A) + 2.0",
		validate: func(t *testing.T, ts []token) {
			if len(ts) != 6 {
				t.Errorf("Expected length 6, but got: %v", ts)
			}
		},
	},
}

func TestTokenize(t *testing.T) {
	for _, c := range tokenCases {
		t.Run(c.name, func(t *testing.T) {
			tokenizer := newParser(strings.NewReader(c.cell))
			tokenizer.scanCell()
			if len(tokenizer.parseErrors) != 0 {
				t.Errorf("Got parse errors: %v", tokenizer.parseErrors)
			}
			c.validate(t, tokenizer.tokens)
		})
	}
}

type parseTestCase struct {
	name     string
	cell     string
	validate func(t *testing.T, n *node, pe []parseError)
}

func assertNodeEqual(t *testing.T, n *node, k nodeKind, v string) {
	if n.kind != k {
		t.Errorf("Expected node kind %v, but got '%v'", k, n.kind)
	}
	if n.rawValue != v {
		t.Errorf("Expected node value %v, but got '%v'", v, n.rawValue)
	}
}

func assertNodeOperator(t *testing.T, n *node, o operator) {
	if n.kind != NodeKindOperator {
		t.Errorf("Expected node kind Operator, but got %v", n.kind)
	}
	if n.operatorValue != o {
		t.Errorf("Expected operator %v, but ot %v", o, n.operatorValue)
	}
}

var parseCases []parseTestCase = []parseTestCase{
	parseTestCase{
		name: "sum column",
		cell: "=sum(A:A) + 2.0",
		validate: func(t *testing.T, n *node, pe []parseError) {
			assertNodeOperator(t, n, Addition)
			assertNodeEqual(t, n.children[1], NodeKindLiteral, "2.0")
			assertNodeEqual(t, n.children[0], NodeKindFunction, "sum")
			assertNodeEqual(t, n.children[0].children[0], NodeKindLiteral, "A:A")
		},
	},
}

func TestParse(t *testing.T) {
	for _, c := range parseCases {
		t.Run(c.name, func(t *testing.T) {
			node, errors, err := Parse(strings.NewReader(c.cell))
			if err != nil {
				t.Errorf("Got error: %v", errors)
			}
			c.validate(t, node, errors)
		})
	}
}
