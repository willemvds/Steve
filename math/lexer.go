package math

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	EOF        = '\uFFFF'
	ADD        = '+'
	MIN        = '-'
	MUL        = '*'
	DIV        = '/'
	OPENBRACE  = '('
	CLOSEBRACE = ')'
)

type ItemType int

const (
	ItemEOF ItemType = iota
	ItemError
	ItemNumber
	ItemLeftBracket
	ItemRightBracket
	ItemOperator
)

var ItemStrings = map[ItemType]string{
	ItemEOF:          "EOF",
	ItemError:        "Error",
	ItemNumber:       "Number",
	ItemLeftBracket:  "Left Parenthesis",
	ItemRightBracket: "Right Parenthesis",
	ItemOperator:     "Operator",
}

type Item struct {
	Typ ItemType
	val string
}

func (i Item) Val() string {
	return i.val
}

func (i Item) String() string {
	str, present := ItemStrings[i.Typ]
	if present {
		return str
	}
	return fmt.Sprintf("Token#%d", i.Typ)
}

type Lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan Item
	state StateFunc
}

func IsSpace(r rune) bool {
	return unicode.IsSpace(r)
}

type StateFunc func(*Lexer) StateFunc

func (l *Lexer) Run() {
	for l.state = lexDefault; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func (l *Lexer) Next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}
	r, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = width
	l.pos += l.width
	return r
}

func (l *Lexer) Emit(t ItemType) {
	l.items <- Item{
		t,
		l.input[l.start:l.pos],
	}
	l.start = l.pos
}

func (l *Lexer) Ignore() {
	l.start = l.pos
}

func (l *Lexer) Backup() {
	l.pos -= l.width
}

func (l *Lexer) Peek() rune {
	r := l.Next()
	defer l.Backup()
	return r
}

func (l *Lexer) Accept(aval string) bool {
	if strings.IndexRune(aval, l.Next()) >= 0 {
		return true
	}
	l.Backup()
	return false
}

func (l *Lexer) AcceptRun(aval string) bool {
	ret := false
	for strings.IndexRune(aval, l.Next()) >= 0 {
		ret = true
	}
	l.Backup()
	return ret
}

func (l *Lexer) Errorf(format string, args ...interface{}) StateFunc {
	l.items <- Item{
		ItemError,
		fmt.Sprintf(format, args),
	}
	return nil
}

func Lex(name, input string) *Lexer {
	l := &Lexer{
		name:  name,
		input: input,
		items: make(chan Item, 2),
	}
	l.state = lexDefault
	//go l.Run()
	return l
}

func (l *Lexer) NextItem() Item {
	for {
		select {
		case item := <-l.items:
			return item
		default:
			if l.state != nil {
				l.state = l.state(l)
			} else {
				close(l.items)
			}
		}
	}
	panic("not reached")
}

func lexDefault(l *Lexer) StateFunc {
	switch r := l.Next(); {
	case IsSpace(r):
		l.Ignore()
	case r == EOF:
		l.Emit(ItemEOF)
		return nil
	case r == OPENBRACE:
		l.Emit(ItemLeftBracket)
	case r == CLOSEBRACE:
		l.Emit(ItemRightBracket)
	case r == ADD || r == MIN || r == MUL || r == DIV:
		l.Emit(ItemOperator)
	case r >= '0' && r <= '9':
		l.Backup()
		return lexNumber
	default:
		return l.Errorf("unrecognized symbol " + string(r))
	}
	return lexDefault
}

func lexNumber(l *Lexer) StateFunc {
	//l.Accept("+-")
	digits := "0123456789"
	l.AcceptRun(digits)
	/*if l.Accept(".") {
		l.AcceptRun(digits)
	}
	if l.Accept("eE") {
		l.Accept("+-")
		l.AcceptRun("0123456789")
	}*/
	l.Emit(ItemNumber)
	return lexDefault
}
