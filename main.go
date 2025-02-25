package main

import (
	"fmt"
	"os"

	"gitlab.com/trum/noteo/internal/app"
)

func main() {
	application := app.New()

	dieOnError("Failed to initialize application", application.Init())
	dieOnError("Failed to run application", application.Run())
}

func dieOnError(msg string, err error) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
