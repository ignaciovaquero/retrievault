package log

import (
	"io"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

var (
	// Msg Is the default instance for logging messages
	Msg = logrus.New()
)

// SetLogLevel Replaces the LogLevel (which defaults to "info")
func SetLogLevel(loglevel string) error {
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		return err
	}
	Msg.Level = level
	return nil
}

// SetOutput Sets the output for the log
func SetOutput(out string) error {
	var file io.Writer
	var err error
	if strings.ToLower(strings.Trim(out, " ")) == "stdout" {
		file = os.Stdout
	} else if strings.ToLower(strings.Trim(out, " ")) == "stderr" {
		file = os.Stderr
	} else {
		file, err = os.OpenFile(out, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	Msg.Out = file
	return nil
}
