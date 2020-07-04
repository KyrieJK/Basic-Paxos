package Basic_Paxos

import (
	"log"
	"time"
)

type proposer struct {
	id           int
	seq          int
	proposeNum   int
	proposeValue string

	acceptors map[int]promise
	nt        network
}

func newProposer(id int, value string, nt network, acceptors ...int) *proposer {
	p := &proposer{
		id:           id,
		seq:          0,
		proposeValue: value,
		acceptors:    make(map[int]promise),
		nt:           nt,
	}

	for _, a := range acceptors {
		p.acceptors[a] = message{}
	}

	return p
}

func (p *proposer) run() {
	var ok bool
	var m message
	log.Println("Proposer start run ... val:", p.proposeValue)
	for !p.majorityReached() {
		if !ok {
			msgs := p.prepare()
			for _, msg := range msgs {
				p.nt.send(msg)
				log.Println("Proposer:send", msg)
			}
		}
		log.Println("Proposer:prepare recv")
		m, ok = p.nt.recv(time.Second)
		if !ok {
			continue
		}

		switch m.typ {
		case Promise:
			p.checkRecvPromise(m)
		default:
			log.Panicf("proposer:%d unexcepted message type:%v", p.id, m.typ)
		}
	}

	log.Printf("proposer: %d promise %d reached quorum %d", p.id, p.getProposeNum(), p.quorum())

	//propose stage
	ms := p.propose()
	for i := range ms {
		p.nt.send(ms[i])
	}
}

func (p *proposer) propose() []message {
	//sendMsgCount := 0
	//var msgList []message
	//for aceptId, aceptMsg := range p.acceptors {
	//	if aceptMsg.seqNumber() == p.getProposeNum() {
	//		msg := message{from: p.id, to: aceptId, typ: Propose, seq: p.getProposeNum(), value: p.proposeValue}
	//		log.Println("Propose val:", msg.value)
	//		msgList = append(msgList, msg)
	//	}
	//	sendMsgCount++
	//	if sendMsgCount > p.quorum() {
	//		break
	//	}
	//}
	//log.Println("proposer propose msg list:", msgList)
	//return msgList
	ms := make([]message, p.quorum())

	i := 0
	for to, promise := range p.acceptors {
		if promise.seqNumber() == p.getProposeNum() {
			ms[i] = message{from: p.id, to: to, typ: Propose, seq: p.getProposeNum(), value: p.proposeValue}
			i++
		}
		if i == p.quorum() {
			break
		}
	}
	return ms
}

func (p *proposer) prepare() []message {
	p.seq++
	sendMsgCount := 0
	//msgList := make([]message, p.quorum())
	//for to, _ := range p.acceptors {
	//	msg := message{
	//		from:  p.id,
	//		to:    to,
	//		typ:   Prepare,
	//		seq:   p.getProposeNum(),
	//		value: p.proposeValue,
	//	}
	//	msgList = append(msgList, msg)
	//	sendMsgCount++
	//	if sendMsgCount == p.quorum() {
	//		break
	//	}
	//}
	ms := make([]message, p.quorum())
	for to := range p.acceptors {
		ms[sendMsgCount] = message{from: p.id, to: to, typ: Prepare, seq: p.getProposeNum()}
		sendMsgCount++
		if sendMsgCount == p.quorum() {
			break
		}
	}

	return ms
}

func (p *proposer) checkRecvPromise(promise message) {
	previousPromise := p.acceptors[promise.from]
	log.Println("prevMsg:", previousPromise, " promiseMsg:", promise)

	if previousPromise.seqNumber() < promise.seqNumber() {
		log.Println("Proposer:", p.id, "get new promise:", promise)
		p.acceptors[promise.from] = promise

		if promise.proposalNumber() > p.proposeNum {
			log.Printf("proposer: %d updated the value [%s] to %s", p.id, p.proposeValue, promise.proposalValue())
			p.proposeNum = promise.proposalNumber()
			p.proposeValue = promise.proposalValue()
		}
	}
}

func (p *proposer) quorum() int {
	return len(p.acceptors)/2 + 1
}

//func (p *proposer) getRecvPromiseCount() int {
//	recvCount := 0
//	for _, aceptMsg := range p.acceptors {
//		if aceptMsg.seqNumber() == p.getProposeNum() {
//			recvCount++
//		}
//	}
//	log.Println("Current proposer recv promise count=", recvCount)
//	return recvCount
//}
//
//func (p *proposer) quorumReached() bool {
//	return p.getRecvPromiseCount() > p.quorum()
//}

func (p *proposer) majorityReached() bool {
	m := 0
	for _, promise := range p.acceptors {
		if promise.seqNumber() == p.getProposeNum() {
			m++
		}
	}
	if m >= p.quorum() {
		return true
	}
	return false
}

func (p *proposer) getProposeNum() int {
	return p.seq<<16 | p.id
}
