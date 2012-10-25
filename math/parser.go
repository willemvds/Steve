package math

import (
	"errors"
	"fmt"
	"strconv"
	"unicode/utf8"
)

type Numnum int

func (n Numnum) Add(v Numnum) Numnum {
	fmt.Printf("<<< Doing %d + %d = %d >>>\n", n, v, n+v)
	return n + v
}

func (n Numnum) Min(v Numnum) Numnum {
	fmt.Printf("<<< Doing %d - %d = %d >>>\n", n, v, n-v)
	return n - v
}

func (n Numnum) Mul(v Numnum) Numnum {
	fmt.Printf("<<< Doing %d * %d = %d >>>\n", n, v, n*v)
	return n * v
}

func (n Numnum) Div(v Numnum) Numnum {
	fmt.Printf("<<< Doing %d / %d = %d >>>\n", n, v, n/v)
	return n / v
}

func (n Numnum) ExecOp(op rune, v Numnum) Numnum {
	switch op {
	case ADD:
		return n.Add(v)
	case MIN:
		return n.Min(v)
	case MUL:
		return n.Mul(v)
	case DIV:
		return n.Div(v)
	}
	panic("this should not happen")
}

type ParseStateFunc func(p *Parser) ParseStateFunc

type Parser struct {
	Name     string
	Tokens   []Item
	state    ParseStateFunc
	parseErr error
	index    int
	res      *Numnum
	op       string
	expr     string
	tree     Tree
}

func (p *Parser) Expr() string {
	return p.expr
}

func (p *Parser) AddToken(t Item) {
	p.Tokens = append(p.Tokens, t)
}

func (p *Parser) BuildTree() (*Tree, error) {
	for p.state = parseNumber; p.state != nil; {
		p.state = p.state(p)
	}
	if p.parseErr != nil {
		return nil, p.parseErr
	}
	return &p.tree, nil
}

func (p *Parser) NextToken() *Item {
	p.index += 1
	if p.index < len(p.Tokens) {
		return &p.Tokens[p.index]
	}
	return nil
}

func (p *Parser) Error(err string) {
	p.parseErr = errors.New(err)
}

func parseNumber(p *Parser) ParseStateFunc {
	token := p.NextToken()
	if token == nil {
		p.Error("Empty token. We're done here")
		return nil
	}
	if token.Typ == ItemLeftBracket {
		p.tree.StackRoot()
		return parseNumber
	}
	if token.Typ != ItemNumber {
		p.Error(fmt.Sprintf("Expected number, got %s", token.String()))
		return nil
	}
	num, err := strconv.Atoi(token.Val())
	n := Numnum(num)
	if err != nil {
		p.Error(fmt.Sprintf("Could not convert %s to a number", num))
		return nil
	}
	p.tree.AddNumber(&n)
	return parseOperator
}

func parseOperator(p *Parser) ParseStateFunc {
	token := p.NextToken()
	if token == nil {
		p.Error("Empty token. We're done here")
		return nil
	}
	if token.Typ == ItemRightBracket {
		p.tree.PopRoot()
		return parseOperator
	}
	if token.Typ != ItemOperator {
		p.Error("expected operator but got something else")
		return nil
	}
	r, _ := utf8.DecodeRuneInString(token.Val())
	p.tree.AddOperator(r)
	return parseNumber
}

func Parse(name string, expr string) (Numnum, error) {
	fmt.Printf("Now parsing %s\n", expr)
	p := &Parser{
		Name:  name,
		expr:  expr,
		index: -1,
		tree:  NewTree(),
	}
	p.Tokens = make([]Item, 0)

	l := Lex(name, expr)
	for next := l.NextItem(); next.Typ != ItemEOF; next = l.NextItem() {
		if next.Typ == ItemError {
			return 0, errors.New("Invalid token encountered, goodbye")
		}
		p.AddToken(next)
	}
	tree, err := p.BuildTree()
	if err != nil {
		return 0, err
	}
	result, err := tree.Parse()
	return result, err
}
