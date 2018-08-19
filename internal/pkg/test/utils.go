package test

import (
	"log"

	"bitbucket.org/airenas/listgo/internal/pkg/messages"
)

type Msg struct {
	M  *messages.QueueMessage
	q  string
	rq string
}

func (m *Msg) equals(o *Msg) bool {
	return m.M.ID == o.M.ID && m.q == o.q && m.rq == o.rq
}

func NewMsg(id string, q string, useRq bool) *Msg {
	rq := ""
	if useRq {
		rq = messages.ResultQueueFor(q)
	}
	return &Msg{M: messages.NewQueueMessage(id), q: q, rq: rq}
}

type Sender struct {
	Msgs []Msg
}

func (sender *Sender) Send(m *messages.QueueMessage, q string, rq string) error {
	log.Printf("Sending msg %s\n", m.ID)
	sender.Msgs = append(sender.Msgs, Msg{m, q, rq})
	return nil
}

func Contains(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}

func ContainsMsg(s []Msg, v *Msg) bool {
	for _, a := range s {
		if a.equals(v) {
			return true
		}
	}
	return false
}
