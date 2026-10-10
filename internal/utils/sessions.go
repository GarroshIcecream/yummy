package utils

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/list"
)

type SessionItem struct {
	Selected          bool
	SessionID         uint
	CreatedAt         time.Time
	UpdatedAt         time.Time
	MessageCount      int
	TotalInputTokens  int
	TotalOutputTokens int
	Summary           string
}

var _ list.Item = &SessionItem{}

func (s SessionItem) Title() string {
	var sessionTitle string
	created := s.CreatedAt.Format("Jan 2, 15:04")
	messageCount := s.MessageCount
	switch messageCount {
	case 0:
		sessionTitle = fmt.Sprintf("Session #%d | %s | no messages", s.SessionID, created)
	case 1:
		sessionTitle = fmt.Sprintf("Session #%d | %s | 1 message", s.SessionID, created)
	default:
		sessionTitle = fmt.Sprintf("Session #%d | %s | %d messages", s.SessionID, created, messageCount)
	}

	if s.Selected {
		sessionTitle = "> " + sessionTitle
	}

	return sessionTitle
}

func (s SessionItem) Description() string {
	if s.MessageCount == 0 {
		return "Empty session"
	}

	var desc string
	if s.Summary != "" {
		summary := s.Summary
		if len(summary) > 100 {
			summary = summary[:97] + "..."
		}
		desc = summary
	} else {
		totalTokens := s.TotalInputTokens + s.TotalOutputTokens
		desc = fmt.Sprintf("Tokens: %d | Last updated: %s", totalTokens, s.UpdatedAt.Format("Jan 2, 15:04"))
	}

	return desc
}

func (s SessionItem) FilterValue() string {
	if s.MessageCount == 0 {
		return "Empty session"
	}

	if s.Summary != "" {
		return s.Summary
	} else {
		return s.Description()
	}
}
