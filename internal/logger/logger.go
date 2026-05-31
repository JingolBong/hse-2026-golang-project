package logger

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	logDirPerm  = 0o755
	logFilePerm = 0o644
)

type Options struct {
	Level   string // "trace"|"debug"|"info"|"warn"|"error"; empty => "info"
	Service string // "connector" | "backend" | "migrator" — emitted as the "service" field
	ToFile  bool   // also duplicate output into logs/ (for local runs); default false in containers
}

type writerHook struct {
	writers   []io.Writer
	levels    []logrus.Level
	formatter logrus.Formatter
}

func (h *writerHook) Levels() []logrus.Level { return h.levels }

func (h *writerHook) Fire(entry *logrus.Entry) error {
	line, err := h.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("format log entry: %w", err)
	}
	for _, w := range h.writers {
		if _, err := w.Write(line); err != nil {
			fmt.Fprintf(os.Stderr, "log hook write error: %v\n", err)
		}
	}
	return nil
}

type fieldHook struct {
	fields logrus.Fields
}

func (h *fieldHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *fieldHook) Fire(entry *logrus.Entry) error {
	for k, v := range h.fields {
		if _, ok := entry.Data[k]; !ok {
			entry.Data[k] = v
		}
	}
	return nil
}

func New(opts Options) *logrus.Logger {
	formatter := &logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "msg",
		},
	}

	log := logrus.New()
	log.SetFormatter(formatter)
	log.SetOutput(os.Stdout)

	lvl, err := logrus.ParseLevel(opts.Level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)

	fields := logrus.Fields{}
	if opts.Service != "" {
		fields["service"] = opts.Service
	}
	if host, err := os.Hostname(); err == nil {
		fields["host"] = host
	}
	if v := os.Getenv("APP_VERSION"); v != "" {
		fields["version"] = v
	}
	log.AddHook(&fieldHook{fields: fields})

	log.AddHook(&writerHook{
		writers:   []io.Writer{os.Stderr},
		levels:    []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
		formatter: formatter,
	})

	if opts.ToFile {
		addFileHooks(log, formatter)
	}

	return log
}

func ToFileEnabled() bool {
	return os.Getenv("LOG_TO_FILE") == "true"
}

func addFileHooks(log *logrus.Logger, formatter logrus.Formatter) {
	if err := os.MkdirAll("logs", logDirPerm); err != nil {
		panic(fmt.Sprintf("create logs dir: %v", err))
	}
	allFile, err := os.OpenFile("logs/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerm)
	if err != nil {
		panic(fmt.Sprintf("open logs.log file: %v", err))
	}
	errFile, err := os.OpenFile("logs/err_logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerm)
	if err != nil {
		panic(fmt.Sprintf("open err_logs.log file: %v", err))
	}

	log.AddHook(&writerHook{
		writers:   []io.Writer{allFile},
		levels:    logrus.AllLevels,
		formatter: formatter,
	})
	log.AddHook(&writerHook{
		writers:   []io.Writer{errFile},
		levels:    []logrus.Level{logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel},
		formatter: formatter,
	})
}
