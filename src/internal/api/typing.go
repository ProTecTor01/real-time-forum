package api

import (
	"encoding/json"

	"src/internal/ws"
)

// typingState belongs to one connection and is only used by its read pump.
type typingState struct {
	hub      *ws.Hub
	userID   int
	username string
	targetID int
}

func (s *typingState) handle(raw []byte) bool {
	var envelope struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		return true
	}
	if envelope.Type == "send_message" {
		s.stop()
		return false
	}
	if envelope.Type != "typing" {
		return false
	}
	var payload struct {
		ToUserID int  `json:"to_user_id"`
		Typing   bool `json:"typing"`
	}
	if json.Unmarshal(envelope.Data, &payload) != nil || payload.ToUserID <= 0 || payload.ToUserID == s.userID {
		return true
	}
	if !payload.Typing {
		if s.targetID == payload.ToUserID {
			s.stop()
		}
		return true
	}
	if s.targetID != payload.ToUserID {
		s.stop()
	}
	s.targetID = payload.ToUserID
	s.send(true)
	return true
}

func (s *typingState) stop() {
	if s.targetID == 0 {
		return
	}
	s.send(false)
	s.targetID = 0
}

func (s *typingState) send(typing bool) {
	s.hub.SendToUser(s.targetID, ws.Event{Type: "typing", Data: map[string]any{
		"from_user_id": s.userID,
		"username":     s.username,
		"typing":       typing,
	}})
}
