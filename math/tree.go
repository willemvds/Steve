package math

import (
	"fmt"
)

type NodeType uint8

const (
	NT_Operator NodeType = iota
	NT_Number
	NT_Blank
)

type Node struct {
	Left   *Node
	Right  *Node
	Parent *Node
	typ    NodeType
	opval  rune
	numval *Numnum
}

func (n *Node) IsOperator() bool {
	return n.typ == NT_Operator
}

func (n *Node) IsBlank() bool {
	return n.typ == NT_Blank
}

func (n *Node) SetType(nt NodeType) {
	n.typ = nt
}

type Tree struct {
	root          *Node
	nextNumParent *Node
	rootStack     []*Node
}

func (t *Tree) StackRoot() {
	t.rootStack = append(t.rootStack, t.root)
	t.root = nil
}

func (t *Tree) PopRoot() {
	oldRoot := t.root
	t.root = t.rootStack[len(t.rootStack)-1]
	t.rootStack = t.rootStack[0 : len(t.rootStack)-1]
	oldRoot.Parent = t.root
	if t.root.Left == nil {
		t.root.Left = oldRoot
	} else {
		t.root.Right = oldRoot
	}
}

func (t *Tree) AddOperator(op rune) {
	if t.root == nil {
		t.root = &Node{
			typ:   NT_Operator,
			opval: op,
		}
		return
	}
	if t.root.IsBlank() {
		t.root.SetType(NT_Operator)
		t.root.opval = op
		return
	}
	if (op == MUL || op == DIV) && t.root.Left != nil {
		nx := &Node{
			typ:    NT_Operator,
			opval:  op,
			Parent: t.root,
		}
		nx.Left = t.root.Right
		t.root.Right = nx
		t.nextNumParent = nx
		return
	}
	t.root.Parent = &Node{
		typ:   NT_Operator,
		opval: op,
		Left:  t.root,
	}
	t.root = t.root.Parent
}

func (t *Tree) AddNumber(num *Numnum) {
	n := &Node{
		typ:    NT_Number,
		numval: num,
	}
	if t.nextNumParent != nil {
		t.nextNumParent.Right = n
		t.nextNumParent = nil
		return
	}
	if t.root == nil {
		t.root = n
		return
	}
	if t.root.Left == nil {
		t.root.Left = n
	} else {
		t.root.Right = n
	}
}

func ReduceNode(n *Node) *Node {
	if n.Left == nil || n.Right == nil {
		fmt.Printf("Cant reduce this bro %s\n", n)
		return n
	}
	fmt.Printf("ReduceNode: %s, Left: %s, Right: %s\n", n.numval, n.Left.numval, n.Right.numval)
	if n.Left.IsOperator() {
		n.Left = ReduceNode(n.Left)
	}
	if n.Right.IsOperator() {
		n.Right = ReduceNode(n.Right)
	}
	if n.Left == nil || n.Right == nil {
		fmt.Printf("Cant reduce this bro, previous reduce turn node to nil\n")
		return n
	}
	result := (*n.Left.numval).ExecOp(n.opval, *n.Right.numval)
	return &Node{
		typ:    NT_Number,
		numval: &result,
		Parent: n.Parent,
	}
}

func (t *Tree) Parse() (Numnum, error) {
	resultNode := ReduceNode(t.root)
	return *resultNode.numval, nil
}

func NewTree() Tree {
	t := Tree{}
	t.rootStack = make([]*Node, 0)
	return t
}
