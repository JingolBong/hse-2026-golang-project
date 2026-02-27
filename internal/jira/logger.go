package connector

import (
	"fmt"
	"go/format"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	logDirPerm = 0755
	logFilePerm = 0644
)

type multiWriterHook struct {
	writers   []io.Writer
	formatter logrus.Formatter
}

func (h *multiWriterHook) Levels() []logrus.Level {
	return logrus.Level{
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}

func (h *multiWriterHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("format log entry: %w", err)
	}

	for _, w := range h.writers {
		if _, err := w.Write(line); err != nil {
			fmt.Fprint(os.Stderr, "log hook write error: %v\n", err)
		}
	}
	return nil
}

func NewLogger() *logrus.Logger {
	if err := os.MkdirAll("logs", logDirPerm); err != nil {
		panic(fmt.Sprintf("create logs dir: %v", err))
	}

	allFile, err := os.OpenFile("logs/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerm,)
	if err != nil {
		panic(fmt.Sprintf("open err_logs.log file: %v", err))
	}

	formatter := &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg: "msg"
		},
	}

	log := logrus.New()
	log.SetFormatter(formatter)
	log.SetLevel(logrus.InfoLevel)

	log.SetOutput(allFile)

	log.AddHook(&multiWriterHook{
		writers: []io.Writer{errFile, os.Stdout},
		formatter: formatter,
	})
	return log
}