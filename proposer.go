package Basic_Paxos

import "log"

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
		id:        id,
		lastSeq:   0,
		value:     value,
		acceptors: make(map[int]promise),
		nt:        nt,
	}

	for _, a := range acceptors {
		p.acceptors[a] = message{}
	}

	return p
}

func (p *proposer) prepare() []message {
	p.seq++
	sendMsgCount := 0
	msgList := make([]message, len(p.acceptors)/2+1)
	i := 0
	for to, _ := range p.acceptors {
		msg := message{
			from:  p.id,
			to:    to,
			typ:   Prepare,
			seq:   p.getProposeNum(),
			value: p.proposeValue,
		}
		msgList = append(msgList, msg)
		sendMsgCount++
		if sendMsgCount > p.quorum() {
			break
		}
	}

	return msgList
}

func (p *proposer) checkRecvPromise(promise message) {
	previousPromise := p.acceptors[promise.from]
	log.Println("prevMsg:", previousPromise, " promiseMsg:", promise)

	if previousPromise.number() < promise.number() {
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

func (p *proposer) getProposeNum() int {
	p.proposeNum = p.seq<<4 | p.id
	return p.proposeNum
}
