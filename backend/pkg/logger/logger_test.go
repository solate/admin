package logger

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
)

func captureStdout(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	b, _ := io.ReadAll(r)
	return string(b)
}

func stripANSI(s string) string {
	re := regexp.MustCompile("\\x1b\\[[0-9;]*m")
	return re.ReplaceAllString(s, "")
}

func TestInitJSONOutput(t *testing.T) {
	out := captureStdout(func() {
		Init(Config{Level: "info", Format: "json"})
		log.Info().Str("key", "value").Msg("Hello")
	})
	cleaned := stripANSI(out)
	if !strings.Contains(cleaned, "Hello") || !strings.Contains(cleaned, "\"key\":\"value\"") {
		t.Fatalf("unexpected json output: %s", out)
	}
}

func TestInitConsoleOutput(t *testing.T) {
	out := captureStdout(func() {
		Init(Config{Level: "info", Format: "console"})
		log.Info().Str("key", "value").Msg("Hello")
	})
	cleaned := stripANSI(out)
	if !strings.Contains(cleaned, "Hello") || !strings.Contains(cleaned, "key=value") {
		t.Fatalf("unexpected console output: %s", out)
	}
}
