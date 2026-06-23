package main

import (
	"log/slog"
	"os"

	"github.com/pkgz/logg"
)

func main() {
	log := logg.New(os.Stdout)
	//log.SetFormat(logg.Json)
	log.SetFlags(logg.LstdFlags)
	//log.DebugMode()

	// Level detected from the message prefix.
	log.Print("DEBUG test logging, but use a somewhat realistic message length.")
	log.Print("INF test logging, but use a somewhat realistic message length.")
	log.Print("ERROR test logging, but use a somewhat realistic message length.")
	log.Print("WRN test logging, but use a somewhat realistic message length.")

	// Level functions.
	log.Info("server started")
	log.Warnf("disk usage at %d%%", 91)

	// Structured fields.
	log.With(logg.F("user", "bob"), logg.F("attempt", 3)).Error("login failed")

	// As a log/slog backend.
	sl := slog.New(log.SlogHandler())
	sl.Info("request handled", "method", "GET", "status", 200)
}
