package message

import (
	_ "fmt"
	
)

type MessageType string

const (
	PREPARE MessageType = "PREPARE"
	ACCEPT MessageType = "ACCEPT"
	PROMISED MessageType = "PROMISED"
	ACCEPTED MessageType = "ACCEPTED"
	REJECTED MessageType = "REJECTED"
)

type messageState string

const (
	IN_TRANSIT messageState = "IN_TRANSIT"
	RECEIVED messageState = "RECEIVED"
)

type PayloadType string

type Message struct {
	state messageState

	Category MessageType
	Proposal *int
	ProposalNumber *int
}

func (m *Message) Send() {
	m.state = IN_TRANSIT
}

func (m *Message) Receive() {
	m.state = RECEIVED
}

func (m *Message) Print() {
	//s := fmt.Sprintf("%v ----> %v(%v: %v) ", m.source, m.Category, m.ProposalNumber, m.Proposal)
	//if m.state == IN_TRANSIT
	//	s = fmt.Sprintf("%v     %v", s, m.destination)
	//else if m.state == RECEIVED
	//	s = fmt.Sprintf("%v ---> %v", s, m.destination)
	//fmt.Println(s)
}
