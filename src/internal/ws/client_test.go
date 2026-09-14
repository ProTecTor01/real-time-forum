package ws

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func echoServer(t *testing.T) (*websocket.Conn, <-chan struct{}) {
	t.Helper()
	hub := NewHub()
	closed := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		client := NewClient(1, hub, conn)
		hub.Register(client)
		go client.WritePump()
		client.ReadPump(client.Send, func() { close(closed) })
	}))
	t.Cleanup(server.Close)
	// Firefox offers this extension. A server that does not negotiate it
	// must still support the connection with ordinary, masked frames.
	conn, response, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), http.Header{
		"Sec-WebSocket-Extensions": {"permessage-deflate"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if response.Header.Get("Sec-WebSocket-Extensions") != "" {
		t.Fatal("unexpected compression negotiation")
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	return conn, closed
}

func TestReadPumpAcceptsMaskedFramesWithCompressionOffer(t *testing.T) {
	conn, _ := echoServer(t)
	for _, size := range []int{1, 125, 126, 1024, 4096, 65536} {
		payload := bytes.Repeat([]byte("a"), size)
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			t.Fatal(err)
		}
		messageType, echoed, err := conn.ReadMessage()
		if err != nil {
			t.Fatal(err)
		}
		if messageType != websocket.TextMessage || !bytes.Equal(echoed, payload) {
			t.Fatalf("wrong echo for %d-byte frame", size)
		}
	}
}

func TestReadPumpRejectsHTTPOnUpgradedConnection(t *testing.T) {
	conn, closed := echoServer(t)
	// The two ASCII bytes 'G', 'E' produce exactly the reported error:
	// RSV1 + opcode 7 in 'G', and a missing client mask in 'E'.
	if _, err := conn.UnderlyingConn().Write([]byte("GET /ws HTTP/1.1\r\n\r\n")); err != nil {
		t.Fatal(err)
	}
	_, _, err := conn.ReadMessage()
	var closeError *websocket.CloseError
	if !errors.As(err, &closeError) || closeError.Code != websocket.CloseProtocolError || closeError.Text != "RSV1 set, bad opcode 7, bad MASK" {
		t.Fatalf("expected protocol error for HTTP bytes, got %v", err)
	}
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("read pump did not close after protocol error")
	}
}
