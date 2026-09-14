package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"src/internal/sessions"
	"src/internal/ws"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
)

func TestTypingWebSocket(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "forum.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema, err := os.ReadFile("../../assets/database/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	sm := sessions.NewSessionManager(db, sessions.DefaultSessionTTL)
	a := NewAPI(db, sm, ws.NewHub())
	server := httptest.NewServer(http.HandlerFunc(a.WS))
	defer server.Close()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	if conn, response, err := websocket.DefaultDialer.Dial(url, nil); err == nil {
		conn.Close()
		t.Fatal("anonymous websocket accepted")
	} else if response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", response)
	}
	names := []string{"Alice", "Bob", "Carol"}
	conns := make([]*websocket.Conn, len(names))
	for i, name := range names {
		_, err := db.Exec(`INSERT INTO users (id, email, username, first_name, last_name, age, gender, password_hash) VALUES (?, ?, ?, 'Test', 'User', 20, 'other', 'unused')`, i+1, fmt.Sprintf("%s@example.com", name), name)
		if err != nil {
			t.Fatal(err)
		}
		session, err := sm.CreateSession(context.Background(), i+1)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := sm.GetSession(context.Background(), session.ID); err != nil {
			t.Fatalf("read test session: %v", err)
		}
		conn, response, err := websocket.DefaultDialer.Dial(url, http.Header{"Cookie": {"session_id=" + session.ID}})
		if err != nil {
			t.Fatalf("connect %s: %v (%v)", name, err, response)
		}
		defer conn.Close()
		conns[i] = conn
	}
	send := func(conn *websocket.Conn, target int, typing bool) {
		t.Helper()
		if err := conn.WriteJSON(ws.Event{Type: "typing", Data: map[string]any{"to_user_id": target, "typing": typing, "username": "Forged", "from_user_id": 999}}); err != nil {
			t.Fatal(err)
		}
	}
	readTyping := func(conn *websocket.Conn, from int, typing bool) {
		t.Helper()
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			var event struct {
				Type string          `json:"type"`
				Data json.RawMessage `json:"data"`
			}
			if err := conn.ReadJSON(&event); err != nil {
				t.Fatal(err)
			}
			if event.Type == "presence" {
				continue
			}
			if event.Type != "typing" {
				t.Fatalf("unexpected event: %s", event.Type)
			}
			var data struct {
				From     int    `json:"from_user_id"`
				Username string `json:"username"`
				Typing   bool   `json:"typing"`
			}
			if err := json.Unmarshal(event.Data, &data); err != nil {
				t.Fatal(err)
			}
			if data.From != from || data.Username != names[from-1] || data.Typing != typing {
				t.Fatalf("unexpected typing: %+v", data)
			}
			return
		}
	}
	send(conns[0], 2, true)
	readTyping(conns[1], 1, true)
	// A refresh must reach the recipient even though the state is already true.
	send(conns[0], 2, true)
	readTyping(conns[1], 1, true)
	send(conns[1], 1, true)
	readTyping(conns[0], 2, true)
	send(conns[1], 1, false)
	readTyping(conns[0], 2, false)
	// Switching recipients clears the previous conversation.
	send(conns[0], 3, true)
	readTyping(conns[1], 1, false)
	readTyping(conns[2], 1, true) // Also proves Carol received no earlier private typing events.
	if err := conns[0].WriteJSON(ws.Event{Type: "send_message", Data: map[string]any{"to_user_id": 3, "body": "Hello"}}); err != nil {
		t.Fatal(err)
	}
	readTyping(conns[2], 1, false)
	var message ws.Event
	if err := conns[2].ReadJSON(&message); err != nil {
		t.Fatal(err)
	}
	if message.Type != "pm_message" {
		t.Fatalf("expected message, got %s", message.Type)
	}
	send(conns[0], 2, true)
	readTyping(conns[1], 1, true)
	conns[0].Close()
	readTyping(conns[1], 1, false)
}

func TestTypingRejectsInvalidRecipients(t *testing.T) {
	s := typingState{hub: ws.NewHub(), userID: 1, username: "Alice"}
	for _, raw := range []string{
		`{"type":"typing","data":{"to_user_id":0,"typing":true}}`,
		`{"type":"typing","data":{"to_user_id":-1,"typing":true}}`,
		`{"type":"typing","data":{"to_user_id":1,"typing":true}}`,
		`{"type":"typing","data":{"to_user_id":"2","typing":true}}`,
		`{"type":"typing","data":{"to_user_id":2,"typing":"yes"}}`,
		`{`,
	} {
		if !s.handle([]byte(raw)) || s.targetID != 0 {
			t.Fatalf("accepted invalid typing: %s", raw)
		}
	}
}
