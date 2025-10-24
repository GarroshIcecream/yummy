package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tmc/langchaingo/llms"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SessionStats struct for session statistics
type SessionStats struct {
	SessionID         uint
	MessageCount      int64
	TotalInputTokens  int64
	TotalOutputTokens int64
}

// Creates new instance of SessionLog struct
func NewSessionLog(dbPath string, opts ...gorm.Option) (*SessionLog, error) {
	dbPath = filepath.Join(dbPath, "session_log.db")
	_, err := os.Stat(dbPath)
	if err != nil {
		log.Printf("Database does not exist at %s, creating new database...", dbPath)
	}

	dbCon, err := gorm.Open(sqlite.Open(dbPath), opts...)
	if err != nil {
		return nil, err
	}

	if err := dbCon.AutoMigrate(GetSessionLogModels()...); err != nil {
		return nil, err
	}

	return &SessionLog{conn: dbCon}, nil
}

// CreateSession creates a new chat session and returns the session ID
func (s *SessionLog) CreateSession() (uint, error) {
	session := SessionHistory{}
	if err := s.conn.Create(&session).Error; err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}
	return session.ID, nil
}

// SaveSessionMessage saves a message to the database
func (s *SessionLog) SaveSessionMessage(sessionID uint, message string, role llms.ChatMessageType, modelName string, content string, inputTokens int, outputTokens int, totalTokens int) error {
	// Convert ChatMessageType to string for database storage
	roleStr := string(role)

	sessionMessage := SessionMessage{
		SessionID:    sessionID,
		Message:      message,
		Role:         roleStr,
		ModelName:    modelName,
		Content:      content,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
	}

	if err := s.conn.Create(&sessionMessage).Error; err != nil {
		return fmt.Errorf("failed to save session message: %w", err)
	}

	return nil
}

// GetSessionMessages retrieves all messages for a given session
func (s *SessionLog) GetSessionMessages(sessionID uint) ([]SessionMessage, error) {
	var messages []SessionMessage
	if err := s.conn.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}
	return messages, nil
}

// GetSessionStats returns statistics for a given session
func (s *SessionLog) GetSessionStats(sessionID uint) (SessionStats, error) {
	var count int64
	var inputTokens, outputTokens int64

	// Count messages
	if err := s.conn.Model(&SessionMessage{}).Where("session_id = ?", sessionID).Count(&count).Error; err != nil {
		return SessionStats{}, fmt.Errorf("failed to count session messages: %w", err)
	}

	// Sum input tokens
	if err := s.conn.Model(&SessionMessage{}).Where("session_id = ?", sessionID).Select("COALESCE(SUM(input_tokens), 0)").Scan(&inputTokens).Error; err != nil {
		return SessionStats{}, fmt.Errorf("failed to sum input tokens: %w", err)
	}

	// Sum output tokens
	if err := s.conn.Model(&SessionMessage{}).Where("session_id = ?", sessionID).Select("COALESCE(SUM(output_tokens), 0)").Scan(&outputTokens).Error; err != nil {
		return SessionStats{}, fmt.Errorf("failed to sum output tokens: %w", err)
	}

	return SessionStats{
		SessionID:         sessionID,
		MessageCount:      count,
		TotalInputTokens:  inputTokens,
		TotalOutputTokens: outputTokens,
	}, nil
}

// GetAllSessions retrieves all chat sessions with their metadata
func (s *SessionLog) GetAllSessions() ([]SessionHistory, error) {
	var sessions []SessionHistory

	// Get all sessions ordered by most recent first
	if err := s.conn.Order("updated_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	return sessions, nil
}

// GetNonEmptySessions retrieves only sessions that have messages
func (s *SessionLog) GetNonEmptySessions() ([]SessionHistory, error) {
	var sessions []SessionHistory

	// Get sessions that have at least one message
	query := `
		SELECT DISTINCT sh.*
		FROM session_histories sh
		INNER JOIN session_messages sm ON sh.id = sm.session_id
		ORDER BY sh.updated_at DESC
	`

	if err := s.conn.Raw(query).Scan(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to get non-empty sessions: %w", err)
	}

	return sessions, nil
}
