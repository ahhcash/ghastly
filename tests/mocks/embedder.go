package mocks

import "github.com/stretchr/testify/mock"

type MockEmbedder struct {
	mock.Mock
}

func (m *MockEmbedder) Embed(text string) ([]float64, error) {
	args := m.Called(text)
	return args.Get(0).([]float64), args.Error(1)
}
