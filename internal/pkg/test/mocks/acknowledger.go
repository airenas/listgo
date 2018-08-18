package mocks

import "github.com/stretchr/testify/mock"

//Acknowledger is a mock
type Acknowledger struct {
	mock.Mock
}

func (m *Acknowledger) Ack(tag uint64, multiple bool) error {
	args := m.Mock.Called()
	return args.Error(0)

}

func (m *Acknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	args := m.Mock.Called()
	return args.Error(0)
}

func (m *Acknowledger) Reject(tag uint64, requeue bool) error {
	args := m.Mock.Called()
	return args.Error(0)
}
