package effe

// Context provides application resources necessary for effe formula evaluation
type Context struct {
	Numbers NumberProvider
	Ranges  RangeProvider
}

type NumberProvider interface {
	ParseNumber(text string) (Number, error)
	Add(a Number, b Number) (Number, error)
}

type Number interface {
	String() string
}

type Range interface {
	IsSingleValue() bool
}

type RangeProvider interface {
	ParseRange(text string) Range
	Intersect(a Range, b Range) Range
	ImplicitIntersect(a Range, b Range) Range
	Single(a Range) Value
	Values(a Range) <-chan Value
}

type Value struct {
	IsA     ValueKind
	Number  Number
	Logical bool
	Text    string
	Range   Range
	Error   error
}

func NumberValue(n Number) Value {
	return Value{
		IsA:    ValueKindNumber,
		Number: n,
	}
}

func LogicalValue(l bool) Value {
	return Value{
		IsA:     ValueKindLogical,
		Logical: l,
	}
}

func TextValue(t string) Value {
	return Value{
		IsA:  ValueKindText,
		Text: t,
	}
}

func ErrorValue(e error) Value {
	return Value{
		IsA:   ValueKindError,
		Error: e,
	}

}

func RangeValue(r Range) Value {
	return Value{
		IsA:   ValueKindRange,
		Range: r,
	}
}

type ValueKind int

const (
	ValueKindNumber ValueKind = iota
	ValueKindText
	ValueKindLogical
	ValueKindError
	ValueKindRange
)
