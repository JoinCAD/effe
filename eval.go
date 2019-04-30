package effe

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/apd"
)

type formulaError string

const (
	null formulaError = "#NULL"
)

type value struct {
	numberValue *apd.Decimal
	stringValue *string
	rangeValue  rangeSpec
	err         *formulaError
}

func (v value) equals(other value) bool {
	return v.numberValue.Cmp(other.numberValue) == 0
}

func (v value) String() string {
	if v.numberValue != nil {
		return v.numberValue.String()
	}
	if v.stringValue != nil {
		return fmt.Sprintf("'%v'", v.stringValue)
	}
	if v.err != nil {
		return fmt.Sprintf("Error %v", v.err)
	}
	return fmt.Sprintf("Range %v", v.rangeValue)
}

type rangeProvider interface {
	iterate(r rangeSpec, c func(v value) error) error
}

type formulaImplementation func(rangeProvider, []value) (value, error)
type session struct {
	rp        rangeProvider
	functions map[string]formulaImplementation
}

func (s session) eval(n node) (value, error) {
	switch n.typ {
	case Function:
		if f, ok := s.functions[strings.ToLower(n.value)]; ok {
			args := []value{}
			for _, c := range n.children {
				v, err := s.eval(c)
				if err != nil {
					return value{}, err
				}
				args = append(args, v)
			}
			return f(s.rp, args)
		}

		return value{}, fmt.Errorf("Unknown function")
	case Range:
		return value{
			rangeValue: n.rangeSpec,
		}, nil
	case Literal:
		// TODO: don't discard condition
		d, _, err := apd.BaseContext.NewFromString(n.value)
		if err != nil {
			return value{
				stringValue: &n.value,
			}, nil
		}
		// TODO: handle condition
		// TODO:logical values (true, false)
		return value{
			numberValue: d,
		}, nil
	default:
		panic("unknown node type")
	}
	// We should never get here.
	return value{}, nil
}
