package mocks

import "github.com/stretchr/testify/mock"

//ResultSaver is a mock
type ResultSaver struct {
	mock.Mock
}

//Save is a mocked Save function
func (m *ResultSaver) Save(ID string, result string) error {
	args := m.Mock.Called(ID, result)
	return args.Error(0)

}
