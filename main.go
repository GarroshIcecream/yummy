package main

import (
	"log/slog"

	"github.com/GarroshIcecream/yummy/yummy/cmd"
	"github.com/GarroshIcecream/yummy/yummy/log"
)

func main() {

	defer log.RecoverPanic("main", func() {
		slog.Error("Application terminated due to unhandled panic")
	})
	
	cmd.Execute()
}
