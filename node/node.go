package node

import (
	"fmt"
	"encoding/json"
	
	"paxos-simulator/message"
)

type nodeState string

const (
	UP nodeState = "UP"
	DOWN nodeState = "DOWN"
)

type queuedMessage struct {
	Sender *Node
	Message *message.Message
}

type Node struct {
	State nodeState
	InboundMessages []queuedMessage
	OutboundMessages []message.Message

	AcceptedProposal *int
	AcceptedProposalNumber *int

	Acceptor *acceptorRole
	Proposer *proposerRole
	
	NumReplicas int
}

type acceptorRole struct {
	HighestProposalNumber int
	HighestProposalNumberProposer *Node
}

type proposerRole struct {
	ProposalNumber int
	PromisedCount int
	AcceptedCount int

	Aborted bool
	HighestProposalNumberInResponses *int
	ProposalCorrespondingToHighestProposalNumberInResponses *int
}


func (n *Node) ReceiveMessage() (*queuedMessage, error) {
	if n.State == DOWN {
		return nil, fmt.Errorf("node is DOWN")
	}
	if len(n.InboundMessages) == 0 {
		return nil, nil
	}
	m := n.InboundMessages[0]
	n.InboundMessages = n.InboundMessages[1:]
	m.Message.Receive()
	//m.Print()
	return &m, nil
}

func (n *Node) ProcessMessageAndConstructResponse(m *queuedMessage) *message.Message {
	var resp *message.Message = nil 
	switch m.Message.Category {
		case message.PREPARE:
			if n.checkUpdateHighestProposal(m.Sender, *m.Message.ProposalNumber) {
				resp = n.constructPromisedMessage()
			} else {
				resp = n.constructRejectedMessage()
			}
		case message.ACCEPT:
			if n.checkAcceptProposal(*m.Message.ProposalNumber, *m.Message.Proposal) {
				resp = n.constructAcceptedMessage()
			} else {
				resp = n.constructRejectedMessage()
			}
		case message.PROMISED:
			if n.checkQuorumPromise(m.Message.ProposalNumber, m.Message.Proposal) {
				resp = n.constructAcceptMessage()
			}
		case message.ACCEPTED:
			n.checkQuorumAccepted()
		case message.REJECTED:
			n.checkRejection()
	}
	if resp != nil {
		n.OutboundMessages = append(n.OutboundMessages, *resp)
	}
	return resp
}

func (n *Node) constructPromisedMessage() *message.Message {
	return &message.Message {
		Category: message.PROMISED,
		Proposal: n.AcceptedProposal,
		ProposalNumber: n.AcceptedProposalNumber,
	}
}

func (n *Node) constructAcceptedMessage() *message.Message {
	return &message.Message {
		Category: message.ACCEPTED,
		Proposal: n.AcceptedProposal,
		ProposalNumber: n.AcceptedProposalNumber,
	}
}

func (n *Node) constructRejectedMessage() *message.Message {
	return &message.Message {
		Category: message.REJECTED,
		ProposalNumber: &n.Acceptor.HighestProposalNumber,
	}
}

func (n *Node) checkUpdateHighestProposal(proposer *Node, proposalNumber int) bool {
	if n.Acceptor == nil {
		n.Acceptor = &acceptorRole{
			HighestProposalNumber: proposalNumber,
			HighestProposalNumberProposer: proposer,
		}
		return true
	} 
	if n.Acceptor.HighestProposalNumber < proposalNumber {
		n.Acceptor.HighestProposalNumber = proposalNumber
		n.Acceptor.HighestProposalNumberProposer = proposer
		return true
	}
	return false
}

func (n *Node) checkAcceptProposal(proposalNumber, proposal int) bool {
	if n.Acceptor == nil || n.Acceptor.HighestProposalNumber <= proposalNumber {
		n.AcceptedProposalNumber = &proposalNumber
		n.AcceptedProposal = &proposal
		return true
	}
	return false
}

func (n *Node) checkQuorumPromise(proposalNumber, proposal *int) bool {
	n.Proposer.PromisedCount++
	if proposalNumber != nil {
		if n.Proposer.HighestProposalNumberInResponses == nil || *n.Proposer.HighestProposalNumberInResponses <= *proposalNumber {
			n.Proposer.HighestProposalNumberInResponses = proposalNumber
			n.Proposer.ProposalCorrespondingToHighestProposalNumberInResponses = proposal
		}
	}
	return n.Proposer.PromisedCount > n.NumReplicas / 2 
}

func (n *Node) checkQuorumAccepted() bool {
	n.Proposer.AcceptedCount++
	return n.Proposer.AcceptedCount > n.NumReplicas / 2 
}

func (n *Node) checkRejection() {
	n.Proposer.Aborted = true
}

func (n *Node) ConstructPrepareMessage(proposalNumber int) *message.Message {
	n.Proposer = &proposerRole {
		ProposalNumber: proposalNumber,
		Aborted: false,
	}
	m := message.Message{
		Category: message.PREPARE,
		ProposalNumber: &proposalNumber,
	}
	n.OutboundMessages = append(n.OutboundMessages, m)
	return &m
}

func (n *Node) constructAcceptMessage() *message.Message {
	p := n.Proposer.ProposalCorrespondingToHighestProposalNumberInResponses
	if p == nil {
		p = &n.Proposer.ProposalNumber
	}
	return &message.Message {
		Category: message.ACCEPT,
		ProposalNumber: &n.Proposer.ProposalNumber,
		Proposal: p,
	}
}

func (n *Node) SendMessage(dest *Node, msg *message.Message) {
	msg.Send()
	fmt.Println(len(dest.InboundMessages))
	dest.InboundMessages = append(dest.InboundMessages, queuedMessage{
		Sender: n,
		Message: msg,
	})
	fmt.Println(len(dest.InboundMessages))
}

func (n *Node) Crash() {
	n.State = DOWN
}

func (n *Node) Recover() {
	n.State = UP
}

func (n *Node) Start() {
	n.Recover()
}

func (n *Node) Print() {
	bytes, _ := json.MarshalIndent(n, "", "    ") 
	fmt.Println(string(bytes))
}

