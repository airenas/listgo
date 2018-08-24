package mocks

import "github.com/stretchr/testify/mock"

//Publisher is a mock
type Publisher struct {
	mock.Mock
}

//Publish is a mocked Publish function
func (m *Publisher) Publish(id string, topic string) error {
	args := m.Mock.Called(id, topic)
	return args.Error(0)

}
