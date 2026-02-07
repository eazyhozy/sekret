package keychain

import (
	"fmt"

	gokeyring "github.com/zalando/go-keyring"
)

const serviceName = "sekret"

// Store provides access to the OS keychain for storing and retrieving secrets.
type Store interface {
	Set(name, value string) error
	Get(name string) (string, error)
	Delete(name string) error
}

// OSStore implements Store using the OS keychain via go-keyring.
type OSStore struct{}

func (s *OSStore) Set(name, value string) error {
	if err := gokeyring.Set(serviceName, name, value); err != nil {
		return fmt.Errorf("failed to save key %q to keychain: %w", name, err)
	}
	return nil
}

func (s *OSStore) Get(name string) (string, error) {
	value, err := gokeyring.Get(serviceName, name)
	if err != nil {
		return "", fmt.Errorf("failed to get key %q from keychain: %w", name, err)
	}
	return value, nil
}

func (s *OSStore) Delete(name string) error {
	if err := gokeyring.Delete(serviceName, name); err != nil {
		return fmt.Errorf("failed to delete key %q from keychain: %w", name, err)
	}
	return nil
}

// NewOSStore returns a new OSStore.
func NewOSStore() *OSStore {
	return &OSStore{}
}
