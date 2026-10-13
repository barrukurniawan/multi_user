package tests

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	defer slog.SetDefault(original)

	r := newTestServer()
	for path, level := range map[string]string{"/ok": "INFO", "/private": "WARN", "/panic": "ERROR"} {
		buf.Reset()
		call(r, path)

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		var log map[string]any
		json.Unmarshal([]byte(lines[len(lines)-1]), &log)

		if log["level"] != level {
			t.Fatalf("%s: level %v, seharusnya %s", path, log["level"], level)
		}
		for _, field := range []string{"request_id", "method", "path", "status_code", "latency_ms"} {
			if log[field] == nil {
				t.Fatalf("%s: field %q tidak ada di log: %s", path, field, lines[len(lines)-1])
			}
		}
	}
}
