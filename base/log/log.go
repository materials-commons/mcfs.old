package log

import (
	"github.com/Sirupsen/logrus"
	"io"
)

// Global log variable.
var L = logrus.New()

func init() {
	L.Formatter = new(logrus.JSONFormatter)
}

// SetOutput sets the output for the global logger.
func SetOutput(out io.Writer) {
	L.Out = out
}

// Field is a convenience function to create a single field that is used in a file or package.
func Field(name, value string) logrus.Fields {
	return logrus.Fields{
		name: value,
	}
}
