package log

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func getOutput(logOutput string) (io.Writer, logrus.Hook, error) {
	var output io.Writer
	var hook logrus.Hook

	switch logOutput {
	case outputStdout:
		output = os.Stdout
	case outputStderr:
		output = os.Stderr
	default:
		output = os.Stderr
	}

	return output, hook, nil
}
