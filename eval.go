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

type value interface {
	implementsCellValue() bool
}

type numberValue struct {
	v *apd.Decimal
}

func (v numberValue) implementsCellValue() bool {
	return true
}

type stringValue struct {
	s string
}

func (v stringValue) implementsCellValue() bool {
	return true
}

type nullValue struct {
}

func (v nullValue) implementsCellValue() bool {
	return true
}

type errValue struct {
	e formulaError
}

func (v errValue) implementsCellValue() bool {
	return true
}

func (v numberValue) String() string {
	return v.v.String()
}

func (v stringValue) String() string {
	return fmt.Sprintf("'%v'", v.s)
}

func equals(v1 value, v2 value) bool {
	switch v1c := v1.(type) {
	case numberValue:
		switch v2c := v2.(type) {
		case numberValue:
			return v1c.v.Cmp(v2c.v) == 0
		}
	case stringValue:
		switch v2c := v2.(type) {
		case stringValue:
			return v1c.s == v2c.s
		}
	case nullValue:
		switch v2.(type) {
		case nullValue:
			return true
		}
	case errValue:
		switch v2c := v2.(type) {
		case errValue:
			return v1c.e == v2c.e
		}
	}

	return false
}

type rangeProvider interface {
	iterate(r rangeSpec, c func(v value) error) error
}

type formulaImplementation func(rangeProvider, []value) (value, error)
type session struct {
	rp        rangeProvider
	functions map[string]formulaImplementation
}

func (s session) eval(n *node) (value, error) {
	switch n.kind {
	case NodeKindFunction:
		if f, ok := s.functions[strings.ToLower(n.value)]; ok {
			args := []value{}
			for _, c := range n.children {
				v, err := s.eval(c)
				if err != nil {
					return nullValue{}, err
				}
				args = append(args, v)
			}
			return f(s.rp, args)
		}

		return nullValue{}, fmt.Errorf("Unknown function: '%v'", strings.ToLower(n.value))

	case NodeKindLiteral:
		// TODO: ranges
		// TODO: don't discard condition
		d, _, err := apd.BaseContext.NewFromString(n.value)
		if err != nil {
			return stringValue{
				s: n.value,
			}, nil
		}
		// TODO: handle condition
		// TODO:logical values (true, false)
		return numberValue{
			v: d,
		}, nil
	default:
		panic("unknown node type")
	}
	// We should never get here.
	return nullValue{}, nil
}
