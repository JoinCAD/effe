package effe

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cockroachdb/apd"
)

type evalTestCase struct {
	formula               string
	expected              value
	expectedErrorFragment string
}

// Stubbed range implements a 10x10 range where each cell contains
// a number determined by position: column# * 10 + row# * 1
type stubbedRange struct{}

var tenByTenRange = makeRangeSpec(1, 10, 1, 10)

func (s stubbedRange) iterate(r rangeSpec, closure func(v value) error) error {
	concreteRange := tenByTenRange.intersect(r)
	for c := concreteRange.cLo; c <= concreteRange.cHi; c++ {
		for r := concreteRange.rLo; r <= concreteRange.rHi; r++ {
			v := apd.New(int64(10*c+r), 0)
			err := closure(numberValue{v: v})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var cases []evalTestCase = []evalTestCase{
	{
		formula:  "sum(A1:A10)",
		expected: numberValue{v: apd.New(560, 0)},
	},
}

func TestEval(t *testing.T) {
	s := session{
		rp: stubbedRange{},
		functions: map[string]formulaImplementation{
			"sum": sum,
		},
	}
	for _, c := range cases {
		n, _, err := Parse(strings.NewReader(c.formula))
		if err != nil {
			t.Fatalf("Error parsing: %v", err)
		}
		result, err := s.eval(n)
		if err != nil {
			t.Fatalf("Error running: %v", err)
		}
		fmt.Println("Got", result, c.expected)
		if !equals(result, c.expected) {
			t.Fatalf("Expected %v, but got %v", c.expected, result)
		}
	}

}
