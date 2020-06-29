package Basic_Paxos

import (
	"log"
	"time"
)

func newAcceptor(id int, nt network, learners ...int) *acceptor {
	newAcceptor := acceptor{id: id, nt: nt}
	newAcceptor.learners = learners
	return &newAcceptor
}

type acceptor struct {
	id       int
	learners []int

	acceptMsg  message
	promiseMsg message
	nt         network
}

func (a *acceptor) run() {
	for {
		m, ok := a.nt.recv(time.Second)
		if !ok {
			continue
		}

		switch m.typ {
		case Prepare:
			promiseMsg := a.recvPrepare(m)
			a.nt.send(*promiseMsg)
			continue
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
		default:
			log.Panicf("acceptor: %d unexcepted message type: %v", a.id, m.typ)
		}
	}
}

func (a *acceptor) recvPrepare(prepare message) *message {
	if a.promiseMsg.seqNumber() >= prepare.seqNumber() {
		log.Println("acceptor ID:", a.id, "Already accept bigger one")
		return nil
	}

	log.Println("acceptor ID:", a.id, " Promise")
	prepare.to = prepare.from
	prepare.from = a.id
	prepare.typ = Promise
	a.acceptMsg = prepare
	return &prepare
}

func (a *acceptor) recvPropose(proposeMsg message) bool {
	if a.acceptMsg.seqNumber() > proposeMsg.seqNumber() || a.acceptMsg.seqNumber() < proposeMsg.seqNumber() {
		return false
	}
	a.acceptMsg = proposeMsg
	a.acceptMsg.typ = Accept
	return true
}
