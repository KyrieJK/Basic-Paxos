package Basic_Paxos

type msgType int

/**
定义message类型，Prepare、Propose分别代表Proposers的准备和提出阶段
Promise、Accept代表Acceptors在两个阶段的响应消息类型
*/
const (
	Prepare msgType = iota + 1
	Propose
	Promise
	Accept
)

/**
定义message结构
*/
type message struct {
	from, to int
	typ      msgType
	n        int
	prevn    int
	value    string
}

func (m message) number() int {
	return m.n
}

func (m message) proposalValue() string {
	switch m.typ {
	case Promise, Accept:
		return m.value
	default:
		panic("unexpected proposalValue")
	}
}

func (m message) proposalNumber() int {
	switch m.typ {
	case Promise:
		return m.prevn
	case Accept:
		return m.n
	default:
		panic("unexpected proposalN")
	}
}

type promise interface {
	number() int
}

type accept interface {
	proposalValue() string
	proposalNumber() int
}
