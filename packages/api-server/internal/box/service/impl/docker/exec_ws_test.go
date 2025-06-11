package docker_test

import (
	"context"
	"testing"

	"github.com/babelcloud/gbox/packages/api-server/internal/box/service/impl/docker"
	model "github.com/babelcloud/gbox/packages/api-server/pkg/box"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestExecWS(t *testing.T) {
	svc := &docker.Service{}
	ctx := context.Background()
	id := "test-container"
	params := &model.BoxExecWSParams{
		TTY:        true,
		WorkingDir: "/app",
		Cmd:        []string{"bash"},
		Args:       []string{"-c", "echo hello"},
	}
	wsConn := &websocket.Conn{}

	result, err := svc.ExecWS(ctx, id, params, wsConn)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

type wsWriter struct {
	conn websocketConn
}

type websocketConn interface {
	WriteMessage(messageType int, data []byte) error
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	err = w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

type mockWebsocketConn struct {
	writeCalled bool
	writeError  error
}

func (m *mockWebsocketConn) WriteMessage(messageType int, data []byte) error {
	m.writeCalled = true
	return m.writeError
}

func TestWsWriter(t *testing.T) {
	t.Run("successful write", func(t *testing.T) {
		mockConn := &mockWebsocketConn{}
		writer := &wsWriter{conn: mockConn}

		data := []byte("test data")
		n, err := writer.Write(data)

		assert.NoError(t, err)
		assert.Equal(t, len(data), n)
		assert.True(t, mockConn.writeCalled)
	})

	t.Run("write error", func(t *testing.T) {
		expectedErr := assert.AnError
		mockConn := &mockWebsocketConn{writeError: expectedErr}
		writer := &wsWriter{conn: mockConn}

		data := []byte("test data")
		n, err := writer.Write(data)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, 0, n)
		assert.True(t, mockConn.writeCalled)
	})
}
