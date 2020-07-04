package Basic_Paxos

import (
	"log"
	"time"
)

func newAcceptor(id int, nt network, learners ...int) *acceptor {
	newAcceptor := acceptor{id: id, nt: nt, promiseMsg: message{}}
	newAcceptor.learners = learners
	return &newAcceptor
}

type acceptor struct {
	id       int
	learners []int

	acceptMsg  message
	promiseMsg promise
	nt         network
}

func (a *acceptor) run() {
	for {
		m, ok := a.nt.recv(time.Hour)
		if !ok {
			continue
		}
		switch m.typ {
		case Propose:
			accepted := a.recvPropose(m)
			if accepted {
				for _, l := range a.learners {
					m := a.acceptMsg
					m.from = a.id
					m.to = l
					m.typ = Accept
					a.nt.send(m)
				}
			}
		case Prepare:
			promiseMsg, ok := a.recvPrepare(m)
			if ok {
				a.nt.send(promiseMsg)
			}
		default:
			log.Panicf("acceptor: %d unexcepted message type: %v", a.id, m.typ)
		}
	}
}

func (a *acceptor) recvPrepare(prepare message) (message, bool) {
	if a.promiseMsg.seqNumber() >= prepare.seqNumber() {
		log.Println("acceptor ID:", a.id, "Already accept bigger one")
		return message{}, false
	}

	//log.Println("acceptor ID:", a.id, " Promise")
	log.Printf("acceptor: %d [promised: %+v] promised %+v", a.id, a.promiseMsg, prepare)
	a.promiseMsg = prepare
	m := message{
		from:    a.id,
		to:      prepare.from,
		typ:     Promise,
		seq:     a.promiseMsg.seqNumber(),
		prevSeq: a.acceptMsg.seq,
		value:   a.acceptMsg.value,
	}
	return m, true
}

func (a *acceptor) recvPropose(proposeMsg message) bool {
	if a.promiseMsg.seqNumber() > proposeMsg.seqNumber() {
		log.Printf("acceptor:%d [promised:%+v] ignored proposal %+v", a.id, a.promiseMsg, proposeMsg)
		return false
	}

	if a.promiseMsg.seqNumber() < proposeMsg.seqNumber() {
		log.Panicf("acceptor: %d [promised:%+v,accept:%+v] accepted proposal %+v", a.id, a.promiseMsg, a.acceptMsg, proposeMsg)
	}
	a.acceptMsg = proposeMsg
	a.acceptMsg.typ = Accept
	return true
}
