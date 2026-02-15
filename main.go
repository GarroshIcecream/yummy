package main

import (
	"log/slog"

	"github.com/GarroshIcecream/yummy/internal/cmd"
	"github.com/GarroshIcecream/yummy/internal/log"
)

func main() {

	defer log.RecoverPanic("main", func() {
		slog.Error("Application terminated due to unhandled panic")
	})

	cmd.Execute()
}
