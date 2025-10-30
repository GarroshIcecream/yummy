package log

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// slogGormLogger implements gorm.io/gorm/logger.Interface to use slog for GORM logging
// This ensures all GORM logs go to the file via slog setup, not stdout
type GormLogger struct {
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
	level                     gormlogger.LogLevel
}

func NewGormLogger(slowThreshold time.Duration, ignoreRecordNotFoundError bool, level gormlogger.LogLevel) *GormLogger {
	return &GormLogger{
		slowThreshold:             slowThreshold,
		ignoreRecordNotFoundError: ignoreRecordNotFoundError,
		level:                     level,
	}
}

// LogMode returns a new logger instance with the specified log level
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info logs info messages using slog
func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Info {
		slog.Info(fmt.Sprintf(msg, data...))
	}
}

// Warn logs warn messages using slog
func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Warn {
		slog.Warn(fmt.Sprintf(msg, data...))
	}
}

// Error logs error messages using slog
func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= gormlogger.Error {
		slog.Error(fmt.Sprintf(msg, data...))
	}
}

// Trace logs SQL queries using slog with structured logging
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.level >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		slog.Error("GORM query error",
			"error", err,
			"sql", sql,
			"rows", rows,
			"duration_ms", elapsed.Milliseconds(),
		)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		slog.Warn("GORM slow query",
			"sql", sql,
			"rows", rows,
			"duration_ms", elapsed.Milliseconds(),
			"threshold_ms", l.slowThreshold.Milliseconds(),
		)
	case l.level == gormlogger.Info:
		slog.Info("GORM query",
			"sql", sql,
			"rows", rows,
			"duration_ms", elapsed.Milliseconds(),
		)
	default:
		slog.Debug("GORM query",
			"sql", sql,
			"rows", rows,
			"duration_ms", elapsed.Milliseconds(),
		)
	}
}
