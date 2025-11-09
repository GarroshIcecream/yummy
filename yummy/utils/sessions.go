package utils

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

type SessionItem struct {
	Selected          bool
	SessionID         uint
	CreatedAt         time.Time
	UpdatedAt         time.Time
	MessageCount      int
	TotalInputTokens  int
	TotalOutputTokens int
}

var _ list.Item = &SessionItem{}

func (s SessionItem) Title() string {
	var sessionTitle string
	created := s.CreatedAt.Format("Jan 2, 15:04")
	messageCount := s.MessageCount
	if messageCount == 0 {
		sessionTitle = fmt.Sprintf("Session #%d - %s (No messages)", s.SessionID, created)
	} else {
		sessionTitle = fmt.Sprintf("Session #%d - %s (%d messages)", s.SessionID, created, messageCount)
	}

	if s.Selected {
		sessionTitle = "🔥 " + sessionTitle
	}

	return sessionTitle
}

func (s SessionItem) Description() string {
	if s.MessageCount == 0 {
		return "Empty session"
	}
	totalTokens := s.TotalInputTokens + s.TotalOutputTokens
	return fmt.Sprintf("Tokens: %d | Last updated: %s", totalTokens, s.UpdatedAt.Format("Jan 2, 15:04"))
}

func (s SessionItem) FilterValue() string {
	return fmt.Sprintf("session %d", s.SessionID)
}
