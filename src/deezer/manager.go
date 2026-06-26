package deezer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/birabittoh/miri"
)

type Manager struct {
	mu       sync.Mutex
	arl      string
	email    string
	password string
	logger   *slog.Logger
}

func NewManager(logger *slog.Logger, arl, email, password string) *Manager {
	return &Manager{
		arl:      arl,
		email:    email,
		password: password,
		logger:   logger,
	}
}

func (m *Manager) CanRenew() bool {
	return m.email != "" && m.password != ""
}

func (m *Manager) ARL() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.arl
}

func (m *Manager) Renew() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.CanRenew() {
		return "", fmt.Errorf("no deezer credentials configured")
	}

	m.logger.Info("renewing ARL cookie via playwright login...")
	arl, err := Login(m.logger, m.email, m.password)
	if err != nil {
		return "", fmt.Errorf("ARL renewal failed: %w", err)
	}

	m.arl = arl
	m.logger.Info("ARL cookie renewed successfully")
	return arl, nil
}

func (m *Manager) EnsureARL() (string, error) {
	arl := m.ARL()
	if arl != "" {
		return arl, nil
	}
	return m.Renew()
}

func IsARLExpiredError(err error) bool {
	return errors.Is(err, miri.ErrInvalidARL)
}

func (m *Manager) ARLExpiredCallback(ctx context.Context) (string, error) {
	return m.Renew()
}
