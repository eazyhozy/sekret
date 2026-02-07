package keychain

import "fmt"

// MockStore implements Store using an in-memory map for testing.
type MockStore struct {
	data map[string]string
}

func (m *MockStore) Set(name, value string) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[name] = value
	return nil
}

func (m *MockStore) Get(name string) (string, error) {
	if m.data != nil {
		if v, ok := m.data[name]; ok {
			return v, nil
		}
	}
	return "", fmt.Errorf("failed to get key %q from keychain: not found", name)
}

func (m *MockStore) Delete(name string) error {
	if m.data != nil {
		if _, ok := m.data[name]; ok {
			delete(m.data, name)
			return nil
		}
	}
	return fmt.Errorf("failed to delete key %q from keychain: not found", name)
}

// NewMockStore returns a new MockStore for testing.
func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]string)}
}
