package testutils

import "github.com/stretchr/testify/mock"

// MockNotifier is a mock implementation of Notifier.
type MockNotifier struct {
	mock.Mock
}

// Notify is a mock implementation of the the real thing.
func (m *MockNotifier) Notify(err error, extras ...interface{}) error {
	return m.Called(err, extras).Error(0)
}

// AutoNotify is a mock implementation of the the real thing.
func (m *MockNotifier) AutoNotify(extras ...interface{}) {
	m.Called(extras)
}
