package effe

import (
	"io"
	"unicode"
)

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
	location uint
	message  string
}

func (pe *parseError) Error() string {
	return pe.message
}

func Parse(r io.RuneScanner) (node, []parseError, error) {
	t := newTokenizer(r)
	t.scanCell()

	return node{}, []parseError{}, nil
}

const (
	// Token type
	TokenTypeNoop     = "Noop"
	TokenTypeOperand  = "Operand"
	TokenTypeFunction = "Function"
	TokenTypeOperator = "Operator"
	TokenTypeSeprator = "Seperator"
	TokenTypeArgument = "Argument"
	TokenTypeUnknown  = "Unknown"
	// Parens
	TokenTypeOpen    = "Open"
	TokenTypeClose   = "Close"
	TokenTypeText    = "Text"
	TokenTypeNumber  = "Number"
	TokenTypeLogical = "Logical"
	TokenTypeError   = "Error"
	TokenTypeRange   = "Range"
)

type token struct {
	value string
	typ   string
}

type tokenizer struct {
	count  uint
	r      io.RuneScanner
	tokens []token
	parseErrors []parseError
}

func newTokenizer(r io.RuneScanner) tokenizer {
	return tokenizer{
		count:       0,
		r:           r,
		tokens:      []token{},
		parseErrors: []parseError{},
	}
}

func (t *tokenizer) read() (r rune, cont bool) {
	c, _, err := t.r.ReadRune()
	t.count = t.count + 1
	if err != nil && err != io.EOF {
		t.parseErrors = append(t.parseErrors, parseError{t.count, "unexpected error reading: " + err.Error()})
		return 0, false
	}
	return c, true
}

func (t *tokenizer) unread() {
	err := t.r.UnreadRune()
	t.count = t.count - 1
	if err != nil {
		t.parseErrors = append(t.parseErrors, parseError{t.count, "unexpected error unreading: " + err.Error()})
	}
}

func (t *tokenizer) accumulateToken(v string, typ string) {
	t.tokens = append(t.tokens, token{
		value: v,
		typ:   typ,
	})
}

func (t *tokenizer) scanCell() {
	r, _ := t.read()
	if r != '=' {
		s := []rune{}
		for r, ok := t.read(); ok; {
			s = append(s, r)
		}
		// TODO parse numbers
		t.accumulateToken(string(s), TokenTypeText)
	} else {
		t.scanFormula()
	}
}

func (t *tokenizer) scanFormula() {
	for t.scanFormulaToken() {
	}
}

func (t *tokenizer) consumeWhiteSpace() bool {
	var r rune
	for r, cont := t.read(); r == ' ' || r == '\t'; {
		if !cont {
			return false
		}
		r, cont := t.read()
	}
	if r != ' ' && r != '\t' {
		t.unread()
		return true
	}
	return true
}

func (t *tokenizer) scanRepeated(predicate func(rune) bool) (string, bool) {
	runes := []rune{}
	cont := true
	for cont {
		r, cont := t.read()
		if predicate(r) {
			runes = append(runes, r)
		} else {
			break
		}
	}
	return string(runes), cont
}

func (t *tokenizer) scanCharacters() (string, bool) {
	return t.scanRepeated(unicode.IsLetter)
}

func (t *tokenizer) scanDigits() (string, bool) {
	return t.scanRepeated(unicode.IsDigit)
}

func (t *tokenizer) dropIfPresent(r rune) {
	c, _ := t.read()
	if c == r {
		return
	}
	t.unread()
}

// Expects either 'A1' or '1' leading string.
func (t *tokenizer) scanRangeSecondHalf(leading string) bool {
	t.dropIfPresent('$')
	rest, cont := t.scanCharacters()
	if !cont {
		return false
	}
	leading = leading + rest
	t.dropIfPresent('$')
	rest, cont = t.scanDigits()
	if !cont {
		return false
	}
	leading = leading + rest
	t.accumulateToken(leading, TokenTypeRange)
	return true
}

// given a leading column, scan the rest of a range token
func (t *tokenizer) scanRange(leading string) bool {
	rest, cont := t.scanDigits()
	if !cont {
		return false
	}
	leading = leading + rest
	r, cont := t.read()
	if !cont {
		return false
	}
	if r != ':' {
		t.unread()
		t.accumulateToken(leading, TokenTypeRange)
		return true
	} else {
		return t.scanRangeSecondHalf(leading)
	}
}

func (t *tokenizer) scanFormulaToken() bool {
	if !t.consumeWhiteSpace() {
		return false
	}
	r, cont := t.read()
	if !cont {
		return cont
	}

	switch r {
	case ',':
		t.accumulateToken("", TokenTypeSeprator)
		return true
	case '(':
		t.accumulateToken("", TokenTypeOpen)
		return true
	case ')':
		t.accumulateToken("", TokenTypeClose)
		return true
	case '+':
	case '-':
	case '%':
	case '*':
	case '/':
	case '^':
	case '>':
		t.accumulateToken(string(r), TokenTypeOperator)
		return true
	case '<':
		r, cont = t.read()
		if cont != true {
			t.accumulateToken("<", TokenTypeOperator)
			return cont
		}
		if r == '>' {
			t.accumulateToken("<>", TokenTypeOperator)
			return true
		} else {
			t.unread()
			t.accumulateToken("<", TokenTypeOperator)
			return true
		}
	}

	if unicode.IsLetter(r) || r == '$' {
		// If it's a letter, read all the letters.
		if r != '$' {
			t.unread()
		}
		s, cont := t.scanCharacters()
		if !cont {
			return false
		}
		// This might be a formula or a range
		r, cont = t.read()
		if !cont {
			return false
		}
		if r == '(' {
			t.accumulateToken(s, TokenTypeFunction)
			t.accumulateToken(s, TokenTypeOpen)
			return true
		}
		// If a digit or '$', then a range.
		if unicode.IsDigit(r) || r == '$' {
			if r != '$' {
				t.unread()
			}
			return t.scanRange(s)
		} else if r == ':' {
			return t.scanRangeSecondHalf(s)

		} else {
			t.unread()
			t.accumulateToken(s, TokenTypeUnknown)
			return true
		}
	}

	// Could be a number, or a range.
	if unicode.IsDigit(r) {
		t.unread()
		leading, cont := t.scanDigits()
		if !cont {
			t.accumulateToken(leading, TokenTypeNumber)
			return false
		}
		r, cont := t.read()
		if !cont {
			t.accumulateToken(leading, TokenTypeNumber)
			return false
		}

		if r == ':' {
			t.unread()
			return t.scanRangeSecondHalf(leading)
		} else if r == '.' {
			rest, cont := t.scanDigits()
			t.accumulateToken(leading+"."+rest, TokenTypeNumber)
			return true
		} else {
			t.unread()
			return true
		}

	}
	// Should never get here!
	return true
}

type parser struct {
	position    uint
	tokens      []token
	parseErrors []parseError
}

func newParser(tokens []token) *parser {
	return &parser{
		position:    0,
		tokens:      tokens,
		parseErrors: []parseError{},
	}
}

// func (p *parser) read() (token, bool) {

// }

func (p *parser) unread() {

}

func (p *parser) parse() {

}

var rangeRegex = regexp.MustCompile("^([a-zA-Z]+)?([0-9]+)?:([a-zA-Z]+)?([0-9]+)?$")

func (p *parser) parseFunction() (node, err) {
	return node{}, nil
}

func (p *parser) createLiteralNode(t token) node {
	// switch t.typ {
	// case TokenTypeRange:
	// 	regexp.
	// case TokenTypeNumber:
	// 	return node{
	// 		typ: 
	// 	}
	// }

		
	return node{}
}

func (p *parser) parseExpression() err {
	operatorStack := []token{}
	output := []node
	for t, more := p.read(); more; t, more = p.read() {

	}
	return nil
}
