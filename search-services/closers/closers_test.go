package closers

import (
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockCloser struct {
	closeErr error
	closed   bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return m.closeErr
}

func TestCloseOrLog_Success(t *testing.T) {
	mc := &mockCloser{}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	CloseOrLog(mc, logger)

	assert.True(t, mc.closed)
}

func TestCloseOrLog_Error(t *testing.T) {
	mc := &mockCloser{closeErr: errors.New("close failed")}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	// Should not panic, just log the error
	CloseOrLog(mc, logger)

	assert.True(t, mc.closed)
}

func TestCloseOrLog_NilLogger(t *testing.T) {
	mc := &mockCloser{closeErr: errors.New("error")}
	CloseOrLog(mc, nil)
	assert.True(t, mc.closed)
}
