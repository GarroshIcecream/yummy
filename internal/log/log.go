package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	initOnce    sync.Once
	initialized atomic.Bool
)

func Setup(logfiledir string, debug bool) {
	logFile := filepath.Join(logfiledir, "logs", "debug.log")
	initOnce.Do(func() {
		logRotator := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    10,
			MaxBackups: 0,
			MaxAge:     30,
			Compress:   false,
		}

		level := slog.LevelInfo
		if debug {
			level = slog.LevelDebug
		}
		logger := slog.NewJSONHandler(logRotator, &slog.HandlerOptions{
			Level:     level,
			AddSource: true,
		})

		slog.SetDefault(slog.New(logger))
		initialized.Store(true)
	})
}

func Initialized() bool {
	return initialized.Load()
}

func RecoverPanic(name string, cleanup func()) {
	if r := recover(); r != nil {
		timestamp := time.Now().Format("20060102-150405")
		filename := fmt.Sprintf("yummy-panic-%s-%s.log", name, timestamp)

		file, err := os.Create(filename)
		if err == nil {
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					fmt.Fprintf(os.Stderr, "Error closing panic log file: %v\n", closeErr)
				}
			}()

			if _, err := fmt.Fprintf(file, "Panic in %s: %v\n\n", name, r); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing panic info: %v\n", err)
			}
			if _, err := fmt.Fprintf(file, "Time: %s\n\n", time.Now().Format(time.RFC3339)); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing panic time: %v\n", err)
			}
			if _, err := fmt.Fprintf(file, "Stack Trace:\n%s\n", debug.Stack()); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing panic stack trace: %v\n", err)
			}

			if cleanup != nil {
				cleanup()
			}
		}
	}
}
