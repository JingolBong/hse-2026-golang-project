package logger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew_LevelParsing(t *testing.T) {
	cases := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"trace":   logrus.TraceLevel,
		"warn":    logrus.WarnLevel,
		"":        logrus.InfoLevel, // empty => info
		"garbage": logrus.InfoLevel, // invalid => info
	}
	for in, want := range cases {
		if got := New(Options{Level: in}).GetLevel(); got != want {
			t.Errorf("level %q: got %v, want %v", in, got, want)
		}
	}
}

func TestNew_InjectsServiceField(t *testing.T) {
	log := New(Options{Level: "info", Service: "connector"})
	var buf bytes.Buffer
	log.SetOutput(&buf) // redirect the main stream so we can inspect it

	log.Info("hello")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("output is not valid JSON: %v (%q)", err, buf.String())
	}
	if entry["service"] != "connector" {
		t.Errorf("service field = %v, want connector", entry["service"])
	}
	for _, k := range []string{"time", "level", "msg"} {
		if _, ok := entry[k]; !ok {
			t.Errorf("missing %q field in %v", k, entry)
		}
	}
}

func TestNew_DebugGating(t *testing.T) {
	// info-level logger must drop debug entries.
	info := New(Options{Level: "info"})
	var infoBuf bytes.Buffer
	info.SetOutput(&infoBuf)
	info.Debug("should be hidden")
	if infoBuf.Len() != 0 {
		t.Errorf("debug leaked at info level: %q", infoBuf.String())
	}

	// debug-level logger must emit debug entries.
	dbg := New(Options{Level: "debug"})
	var dbgBuf bytes.Buffer
	dbg.SetOutput(&dbgBuf)
	dbg.Debug("should appear")
	if dbgBuf.Len() == 0 {
		t.Error("debug entry missing at debug level")
	}
}
