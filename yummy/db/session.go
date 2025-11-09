package db

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/GarroshIcecream/yummy/yummy/config"
	"github.com/GarroshIcecream/yummy/yummy/log"
	utils "github.com/GarroshIcecream/yummy/yummy/utils"
	"github.com/tmc/langchaingo/llms"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// SessionStats struct for session statistics
type SessionStats struct {
	SessionID    uint
	MessageCount int
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// Creates new instance of SessionLog struct
func NewSessionLog(dbPath string, config *config.DatabaseConfig, opts ...gorm.Option) (*SessionLog, error) {
	dbDir := filepath.Join(dbPath, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		slog.Error("Failed to create database directory", "dir", dbDir, "error", err)
		return nil, err
	}

	dbPath = filepath.Join(dbDir, config.SessionLogDBName)
	_, err := os.Stat(dbPath)
	if err != nil {
		slog.Info("Database does not exist at %s, creating new database...", "dbPath", dbPath, "error", err)
	}

	dbCon, err := gorm.Open(sqlite.Open(dbPath), opts...)
	if err != nil {
		return nil, err
	}

	// Configure GORM to use slog logger (logs to file via slog setup, not stdout)
	dbCon.Logger = log.NewGormLogger(200*time.Millisecond, true, gormlogger.Info)

	if err := dbCon.AutoMigrate(GetSessionLogModels()...); err != nil {
		return nil, err
	}

	return &SessionLog{conn: dbCon}, nil
}

// CreateSession creates a new chat session and returns the session ID
func (s *SessionLog) CreateSession() (uint, error) {
	session := SessionHistory{}
	if err := s.conn.Create(&session).Error; err != nil {
		slog.Error("Error creating session", "error", err)
		return 0, err
	}
	return session.ID, nil
}

// SaveSessionMessage saves a message to the database
func (s *SessionLog) SaveSessionMessage(sessionID uint, message string, role llms.ChatMessageType, modelName string, inputTokens int, outputTokens int, totalTokens int) error {
	// Convert ChatMessageType to string for database storage
	roleStr := string(role)

	sessionMessage := SessionMessage{
		SessionID:    sessionID,
		Message:      message,
		Role:         roleStr,
		ModelName:    modelName,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
	}

	if err := s.conn.Create(&sessionMessage).Error; err != nil {
		slog.Error("Error saving session message", "error", err)
		return err
	}

	return nil
}

// GetSessionMessages retrieves all messages for a given session
func (s *SessionLog) GetSessionMessages(sessionID uint) ([]SessionMessage, error) {
	var messages []SessionMessage
	if err := s.conn.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error; err != nil {
		slog.Error("Error getting session messages", "error", err)
		return nil, err
	}
	return messages, nil
}

// GetSessionStats returns statistics for a given session
func (s *SessionLog) GetSessionStats(sessionID uint) (SessionStats, error) {
	var stats SessionStats
	if err := s.conn.Model(&SessionMessage{}).
		Where("session_id = ?", sessionID).
		Select(
			"session_id, "+
				"SUM(CASE WHEN role != ? THEN 1 ELSE 0 END) AS message_count, "+
				"COALESCE(SUM(input_tokens), 0) AS input_tokens, "+
				"COALESCE(SUM(output_tokens), 0) AS output_tokens, "+
				"COALESCE(SUM(total_tokens), 0) AS total_tokens",
			"system",
		).
		Scan(&stats).Error; err != nil {
		slog.Error("Error aggregating session stats", "error", err)
		return stats, err
	}

	return stats, nil
}

// GetAllSessions retrieves all chat sessions with their metadata
func (s *SessionLog) GetAllSessions() ([]SessionHistory, error) {
	var sessions []SessionHistory

	// Get all sessions ordered by most recent first
	if err := s.conn.Order("updated_at DESC").Find(&sessions).Error; err != nil {
		slog.Error("Error getting sessions", "error", err)
		return nil, err
	}

	return sessions, nil
}

// GetNonEmptySessions retrieves only sessions that have messages
func (s *SessionLog) GetNonEmptySessions() ([]*utils.SessionItem, error) {
	var sessions []*utils.SessionItem

	// Get sessions that have at least one message
	query := `
		SELECT DISTINCT sh.id as session_id, sh.created_at, sh.updated_at, COUNT(sm.id) as message_count, SUM(sm.input_tokens) as total_input_tokens, SUM(sm.output_tokens) as total_output_tokens
		FROM session_histories sh
		INNER JOIN session_messages sm ON sh.id = sm.session_id
		GROUP BY sh.id
		ORDER BY sh.updated_at DESC
	`

	if err := s.conn.Raw(query).Scan(&sessions).Error; err != nil {
		slog.Error("Error getting non-empty sessions", "error", err)
		return nil, err
	}

	return sessions, nil
}
