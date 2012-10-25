package math

import (
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
	Name   string
	Tokens []Item
	state  ParseStateFunc
	index  int
	res    *Numnum
	op     string
	expr   string
	tree   Tree
}

func (p *Parser) Expr() string {
	return p.expr
}

func (p *Parser) AddToken(t Item) {
	p.Tokens = append(p.Tokens, t)
}

func (p *Parser) BuildTree() *Tree {
	for p.state = parseNumber; p.state != nil; {
		p.state = p.state(p)
	}
	return &p.tree
}

func (p *Parser) NextToken() *Item {
	p.index += 1
	if p.index < len(p.Tokens) {
		return &p.Tokens[p.index]
	}
	return nil
}

func (p *Parser) Result() Numnum {
	return *p.res
}

func parseNumber(p *Parser) ParseStateFunc {
	token := p.NextToken()
	if token == nil {
		return nil
	}
	if token.Typ == ItemLeftBracket {
		p.tree.StackRoot()
		return parseNumber
	}
	if token.Typ != ItemNumber {
		fmt.Printf("WAT!!")
		return nil
	}
	num, err := strconv.Atoi(token.Val())
	n := Numnum(num)
	if err != nil {
		fmt.Printf("That thing is not a number, ( . Y . )")
	}
	p.tree.AddNumber(&n)
	return parseOperator
}

func parseOperator(p *Parser) ParseStateFunc {
	token := p.NextToken()
	if token == nil {
		return nil
	}
	if token.Typ == ItemRightBracket {
		p.tree.PopRoot()
		return parseOperator
	}
	if token.Typ != ItemOperator {
		fmt.Printf("expecting operator, got nothing\n")
		return nil
	}
	r, _ := utf8.DecodeRuneInString(token.Val())
	p.tree.AddOperator(r)
	return parseNumber
}

func Parse(name string, expr string) Numnum {
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
		p.AddToken(next)
	}
	tree := p.BuildTree()
	result, _ := tree.Parse()
	return result
}
