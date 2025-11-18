package logging

import (
	"log"
	"os"
)

// New returns a plain stdlib logger so you can start wiring features without
// worrying about structured logging just yet.
func New(prefix string) *log.Logger {
	if prefix == "" {
		prefix = "duo"
	}
	return log.New(os.Stdout, prefix+": ", log.LstdFlags)
}
