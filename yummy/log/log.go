package log

import (
	"fmt"
	"log/slog"
	"os"
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

func Setup(logFile string, debug bool) {
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
			defer file.Close()

			fmt.Fprintf(file, "Panic in %s: %v\n\n", name, r)
			fmt.Fprintf(file, "Time: %s\n\n", time.Now().Format(time.RFC3339))
			fmt.Fprintf(file, "Stack Trace:\n%s\n", debug.Stack())

			if cleanup != nil {
				cleanup()
			}
		}
	}
}
