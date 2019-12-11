package effe

import (
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"unicode"
)

type node struct {
	kind          nodeKind
	value         string
	children      []*node
	operatorValue operator
}

type nodeKind int

const (
	NodeKindFunction nodeKind = iota
	NodeKindLiteral
	NodeKindOperator
	NodeKindHole
)

type operator int

const (
	Intersection operator = iota
	UnaryNegation
	Percent
	Exponentiation
	Multiplication
	Division
	Addition
	Subtraction
	Concatenation
	Equality
	GreaterThan
	LessThan
	GreaterThanOrEqual
	LessThanOrEqual
	Inequality
)

type parseError struct {
	location uint
	message  string
}

func (pe *parseError) Error() string {
	return pe.message
}

func Parse(r io.RuneScanner) (*node, []parseError, error) {
	var t *tokenizer
	var p *parser
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Panic parsing: ", err)
			fmt.Println("Tokenizer: ", *t)
			fmt.Println("Parser: ", *p)
			debug.PrintStack()
		}
	}()
	t = newTokenizer(r)
	t.scanCell()
	p = newParser(t.tokens)
	p.parse()
	return p.next[0], t.parseErrors, nil
}

const (
	// Token type
	TokenTypeNoop     = "Noop"
	TokenTypeFunction = "Function"
	TokenTypeOperator = "Operator" // TODO: prefix postfix infix????
	TokenTypeSeprator = "Seperator"
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
	value         string
	typ           string
	operatorValue operator
}

type tokenizer struct {
	count       uint
	r           io.RuneScanner
	tokens      []token
	parseErrors []parseError
}

func newTokenizer(r io.RuneScanner) *tokenizer {
	return &tokenizer{
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
		log.Println("Error reading", err)
		return 0, false
	}
	return c, err != io.EOF
}

func (t *tokenizer) unread() {
	err := t.r.UnreadRune()
	t.count = t.count - 1
	if err != nil {
		t.parseErrors = append(t.parseErrors, parseError{t.count, "unexpected error unreading: " + err.Error()})
	}
}

func (t *tokenizer) accumulateToken(v string, typ string) {
	token := token{
		value: v,
		typ:   typ,
	}

	if typ == TokenTypeOperator {
		token.operatorValue = getOperator(v)
	}

	t.tokens = append(t.tokens, token)
}

func (t *tokenizer) scanCell() {
	r, _ := t.read()
	if r != '=' {
		s := []rune{r}
		for r, ok := t.read(); ok; r, ok = t.read() {
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

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func (t *tokenizer) consumeWhiteSpace() bool {
	var r rune
	var cont bool
	for r, cont = t.read(); isWhitespace(r); r, cont = t.read() {
		if !cont {
			return false
		}
	}
	if cont && !isWhitespace(r) {
		t.unread()
		return true
	}
	return true
}

func (t *tokenizer) scanRepeated(predicate func(rune) bool) string {
	runes := []rune{}
	cont := true
	var r rune
	for cont {
		r, cont = t.read()
		if predicate(r) {
			runes = append(runes, r)
		} else {
			if cont {
				t.unread()
			}
			break
		}
	}
	return string(runes)
}

func (t *tokenizer) scanCharacters() string {
	return t.scanRepeated(unicode.IsLetter)
}

func (t *tokenizer) scanDigits() string {
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
	// TODO: dropping out '$' mean that the parsed representation can be no help in moving formulas around.
	t.dropIfPresent('$')
	rest := t.scanCharacters()
	leading = leading + rest
	t.dropIfPresent('$')
	rest = t.scanDigits()
	leading = leading + rest
	t.accumulateToken(leading, TokenTypeRange)
	return true
}

// given a leading column, scan the rest of a range token (digits)
func (t *tokenizer) scanRange(leading string) bool {
	rest := t.scanDigits()
	leading = leading + rest
	r, cont := t.read()
	if !cont {
		t.accumulateToken(leading, TokenTypeRange)
		return false
	}
	if r != ':' {
		t.unread()
		t.accumulateToken(leading, TokenTypeRange)
		return true
	} else {
		leading = leading + ":"
		return t.scanRangeSecondHalf(leading)
	}
}

func (t *tokenizer) scanFormulaToken() bool {
	if !t.consumeWhiteSpace() {
		fmt.Println("falling from consumeWhitespace")
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
		fallthrough
	case '-':
		fallthrough
	case '%':
		fallthrough
	case '*':
		fallthrough
	case '/':
		fallthrough
	case '^':
		fallthrough
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
		s := t.scanCharacters()

		// This might be a formula or a range
		r, cont = t.read()
		if !cont {
			return false
		}
		if r == '(' {
			t.accumulateToken(s, TokenTypeFunction)
			t.accumulateToken("(", TokenTypeOpen)
			return true
		}
		// If a digit or '$', then a range.
		if unicode.IsDigit(r) || r == '$' {
			if r != '$' {
				t.unread()
			}
			return t.scanRange(s)
		} else if r == ':' {
			s = s + ":"
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
		leading := t.scanDigits()
		r, cont := t.read()
		if !cont {
			// TODO: what about r?
			t.accumulateToken(leading, TokenTypeNumber)
			return false
		}

		if r == ':' {
			t.unread()
			return t.scanRangeSecondHalf(leading)
		} else if r == '.' {
			rest := t.scanDigits()
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
	position      int
	tokens        []token
	parseErrors   []parseError
	operator      []token
	next          []*node
	argCountStack []int
	// Used to distinguish infix and unary minus
	infix bool
}

func newParser(tokens []token) *parser {
	return &parser{
		position:    0,
		tokens:      tokens,
		parseErrors: []parseError{},
	}
}

func (p *parser) more() bool {
	return p.position < len(p.tokens)
}

func (p *parser) read() token {
	var t = p.tokens[p.position]
	p.position++
	return t
}

func (p *parser) unread() {
	if p.position == 0 {
		panic("cannot unread past start of tokens")
	}
	p.position--
}

func (p *parser) peek() token {
	return p.tokens[p.position]
}

func (p *parser) pushOperator(t token) {
	p.operator = append(p.operator, t)
}

func (p *parser) peekOperator() token {
	return p.operator[len(p.operator)-1]
}

func (p *parser) popOperator() {
	p.operator = p.operator[:len(p.operator)-1]
}

func (p *parser) moreOperator() bool {
	return len(p.operator) > 0
}

func getOperator(s string) operator {
	switch s {
	case "-u":
		return UnaryNegation
	case "%":
		return Percent
	case "^":
		return Exponentiation
	case "*":
		return Multiplication
	case "/":
		return Division
	case "+":
		return Addition
	case "-":
		return Subtraction
	case "&":
		return Concatenation
	case "=":
		return Equality
	case ">":
		return GreaterThan
	case "<":
		return LessThan
	case ">=":
		return GreaterThanOrEqual
	case "<=":
		return LessThanOrEqual
	case "<>":
		return Inequality
	default:
		panic("Unkown operator: " + s)
	}
}

func operatorPrecedence(o operator) int {
	switch o {
	case Intersection:
		return 8
	case UnaryNegation:
		return 7
	case Percent:
		return 6
	case Exponentiation:
		return 5
	case Multiplication:
		return 4
	case Division:
		return 4
	case Addition:
		return 3
	case Subtraction:
		return 3
	case Concatenation:
		return 2
	case Equality:
		return 1
	case GreaterThan:
		return 1
	case LessThan:
		return 1
	case GreaterThanOrEqual:
		return 1
	case LessThanOrEqual:
		return 1
	case Inequality:
		return 1
	}
	// should never happen
	return 0
}

func operatorArgs(o operator) int {
	switch o {
	case UnaryNegation:
		return 1
	case Percent:
		return 1
	default:
		return 2
	}
}

func leftAssociative(o operator) bool {
	if o == UnaryNegation {
		return false
	}
	return true
}

// TODO: parse numbers
// TODO: parse ranges
// TODO: parse logical and text?
func buildSimpleNode(t token) *node {
	switch t.typ {
	case TokenTypeNumber:
		return &node{
			kind:  NodeKindLiteral,
			value: t.value,
		}
	case TokenTypeRange:
		return &node{
			kind:  NodeKindLiteral,
			value: t.value,
		}
	case TokenTypeNoop:
		return &node{
			kind: NodeKindHole,
		}
	}
	panic("unexpected token type for simple node" + t.typ)
}

func (p *parser) outputOperator(o operator) {
	var nargs = operatorArgs(o)
	var n = &node{
		kind:          NodeKindOperator,
		operatorValue: o,
	}
	n.children = make([]*node, nargs)
	copy(n.children, p.next[len(p.next)-nargs:])
	p.next = p.next[:len(p.next)-nargs]
	p.next = append(p.next, n)
}

func (p *parser) output(t token) {
	// If the last one was a range (or a formula), implicitly create a ' ' intersection operator
	if t.typ != TokenTypeNoop && t.typ != TokenTypeFunction && len(p.next) > 0 && p.next[len(p.next)-1].kind == NodeKindHole {
		// Fill the hole.
		p.next[len(p.next)-1] = buildSimpleNode(t)
		return
	}

	if t.typ == TokenTypeRange && p.next[len(p.next)-1].kind != NodeKindHole {
		p.next = append(p.next, buildSimpleNode(t))
		p.outputOperator(Intersection)
		return
	}

	if t.typ == TokenTypeRange ||
		t.typ == TokenTypeNumber ||
		t.typ == TokenTypeLogical ||
		t.typ == TokenTypeText ||
		t.typ == TokenTypeNoop {
		p.next = append(p.next, buildSimpleNode(t))
	}

	if t.typ == TokenTypeFunction {
		nargs := p.argCountStack[len(p.argCountStack)-1]
		log.Println("Function", nargs)

		//pop
		p.argCountStack = p.argCountStack[:len(p.argCountStack)-1]

		n := &node{
			kind:  NodeKindFunction,
			value: t.value,
		}
		n.children = make([]*node, nargs)
		copy(n.children, p.next[len(p.next)-nargs:])
		p.next = p.next[:len(p.next)-nargs]
		p.next = append(p.next, n)
	}

	if t.typ == TokenTypeOperator {
		p.outputOperator(t.operatorValue)
	}
}

func (p *parser) parse() {
	// Shunting yard algorithm
	for p.more() {
		t := p.read()
		switch t.typ {
		// Values
		case TokenTypeText:
			fallthrough
		case TokenTypeNumber:
			fallthrough
		case TokenTypeLogical:
			fallthrough
		case TokenTypeError:
			fallthrough
		case TokenTypeRange:
			p.infix = true
			p.output(t)

		case TokenTypeFunction:
			p.infix = false

			if p.peek().typ != TokenTypeOpen {
				// ERROR
			}
			// Drop the open, we'll treat it as being subsumed into the function
			p.read()
			p.pushOperator(t)
			p.argCountStack = append(p.argCountStack, 1)
			// Placeholder in case we get no arguments
			// output needs to special case a 1-arg function with noop arg
			p.output(token{typ: TokenTypeNoop})

		case TokenTypeOperator:

			// Check if subtraction should be converted to unary minus
			if !p.infix && t.operatorValue == Subtraction {
				t.operatorValue = UnaryNegation
			}
			p.infix = false

			for p.moreOperator() {
				var next = p.peekOperator()
				if next.typ == TokenTypeOpen {
					break
				}
				if next.typ == TokenTypeFunction {
					p.output(next)
					p.popOperator()
				}
				if next.typ == TokenTypeOperator {
					if operatorPrecedence(next.operatorValue) > operatorPrecedence(t.operatorValue) ||
						(operatorPrecedence(next.operatorValue) == operatorPrecedence(t.operatorValue) && leftAssociative(next.operatorValue)) {
						p.output(next)
						p.popOperator()
					}
				}

			}

			p.pushOperator(t)
		case TokenTypeSeprator:
			p.infix = false

			// Pop operators until we get to a function
			// increment the most recent
			for {
				o := p.operator[len(p.operator)-1]
				if o.typ == TokenTypeFunction {
					break
				} else {
					p.output(o)
					p.operator = p.operator[:len(p.operator)-2]
				}
			}
			p.argCountStack[len(p.argCountStack)] = p.argCountStack[len(p.argCountStack)] + 1
			p.output(token{typ: TokenTypeNoop})
			// TODO: this breaks for 1 + sum(,B2)
		case TokenTypeOpen:
			p.infix = false

			p.operator = append(p.operator, t)
		case TokenTypeClose:
			p.infix = false

			// Pop until we get to a function or an open.
			// If left paren -- pop and discard
			// If function
			// push to output
			for {
				if len(p.operator) == 0 {
					// ERROR unbalanced parens
				}
				o := p.operator[len(p.operator)-1]
				if o.typ == TokenTypeOpen {
					p.operator = p.operator[:len(p.operator)-2]
					break
				} else if o.typ == TokenTypeFunction {
					// Leave it for something else to pop and push over, or the final cleanup
					break
				} else {
					p.output(o)
					p.operator = p.operator[:len(p.operator)-2]
				}

			}
		default:
			panic("unexected token type:" + t.typ)
		}
	}

	for p.moreOperator() {
		t := p.peekOperator()
		if t.typ == TokenTypeOpen {
			panic("unbalanced parens")
		}
		p.output(t)
		p.popOperator()
	}

	//
}
