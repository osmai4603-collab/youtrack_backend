package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"youtrack_backend/internal/platform/config"
)

func TestDailyFileWriter_WritingAndRotation(t *testing.T) {
	tempDir := t.TempDir()

	currentDate := time.Date(2026, time.September, 28, 10, 0, 0, 0, time.UTC)
	var mu sync.Mutex
	getNow := func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		return currentDate
	}

	w, err := newDailyFileWriterWithClock(tempDir, "app", getNow)
	if err != nil {
		t.Fatalf("newDailyFileWriterWithClock failed: %v", err)
	}

	// Write Day 1
	msg1 := []byte("event from day 1\n")
	if _, err := w.Write(msg1); err != nil {
		t.Fatalf("Write day 1 failed: %v", err)
	}

	day1File := filepath.Join(tempDir, "app-2026-09-28.log")
	if _, err := os.Stat(day1File); err != nil {
		t.Fatalf("expected day 1 log file %q to exist: %v", day1File, err)
	}

	// Rotate to Day 2
	mu.Lock()
	currentDate = time.Date(2026, time.September, 29, 0, 1, 0, 0, time.UTC)
	mu.Unlock()

	msg2 := []byte("event from day 2\n")
	if _, err := w.Write(msg2); err != nil {
		t.Fatalf("Write day 2 failed: %v", err)
	}

	day2File := filepath.Join(tempDir, "app-2026-09-29.log")
	if _, err := os.Stat(day2File); err != nil {
		t.Fatalf("expected day 2 log file %q to exist: %v", day2File, err)
	}

	if err := w.Sync(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Check content of Day 1 file
	b1, err := os.ReadFile(day1File)
	if err != nil {
		t.Fatalf("ReadFile day 1 failed: %v", err)
	}
	if !strings.Contains(string(b1), "event from day 1") {
		t.Errorf("expected day 1 content, got: %s", string(b1))
	}

	// Check content of Day 2 file
	b2, err := os.ReadFile(day2File)
	if err != nil {
		t.Fatalf("ReadFile day 2 failed: %v", err)
	}
	if !strings.Contains(string(b2), "event from day 2") {
		t.Errorf("expected day 2 content, got: %s", string(b2))
	}
}

func TestDailyFileWriter_ConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()

	w, err := newDailyFileWriter(tempDir, "concurrent")
	if err != nil {
		t.Fatalf("newDailyFileWriter failed: %v", err)
	}
	defer w.Close()

	var wg sync.WaitGroup
	workers := 10
	iterations := 20

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				line := fmt.Sprintf("worker %d line %d\n", workerID, j)
				_, _ = w.Write([]byte(line))
			}
		}(i)
	}

	wg.Wait()

	today := time.Now().Format("2006-01-02")
	expectedFile := filepath.Join(tempDir, fmt.Sprintf("concurrent-%s.log", today))
	data, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("expected log file to exist and be readable: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != workers*iterations {
		t.Errorf("expected %d log lines, got %d", workers*iterations, len(lines))
	}
}

func TestPrettyHandler_ConcurrentWrites(t *testing.T) {
	var output strings.Builder
	logger := slog.New(&prettyHandler{w: &output})

	const workers = 10
	const iterations = 20

	var wg sync.WaitGroup
	for workerID := 0; workerID < workers; workerID++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for iteration := 0; iteration < iterations; iteration++ {
				logger.Info(fmt.Sprintf("worker-%d-line-%d", workerID, iteration))
			}
		}(workerID)
	}
	wg.Wait()

	lines := strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n")
	if got, want := len(lines), workers*iterations; got != want {
		t.Fatalf("expected %d complete log lines, got %d", want, got)
	}

	seen := make(map[string]bool, workers*iterations)
	for _, line := range lines {
		start := strings.Index(line, "worker-")
		if start < 0 {
			t.Errorf("log line does not contain a message: %q", line)
			continue
		}
		messageEnd := strings.IndexByte(line[start:], ' ')
		if messageEnd < 0 {
			t.Errorf("log line does not contain the expected padding: %q", line)
			continue
		}
		message := line[start : start+messageEnd]
		if seen[message] {
			t.Errorf("duplicate log message %q", message)
		}
		seen[message] = true
	}
	if got := len(seen); got != workers*iterations {
		t.Errorf("expected %d unique log messages, got %d", workers*iterations, got)
	}
}

func TestNewLogger_DailyLogFilesCreated(t *testing.T) {
	tempDir := t.TempDir()

	settings := config.LoggerSettings{
		Level:      "info",
		ForceColor: false,
		ToFile:     true,
		LogDir:     tempDir,
	}

	log, closer := newLogger(settings)
	if log == nil {
		t.Fatal("expected logger to be non-nil")
	}

	log.Info("sample info event", slog.String("key", "val1"))
	log.Error("sample error event", slog.Int("code", 500))

	if err := closer.Close(); err != nil {
		t.Fatalf("closer.Close() failed: %v", err)
	}

	today := time.Now().Format("2006-01-02")
	appFile := filepath.Join(tempDir, fmt.Sprintf("app-%s.log", today))
	errFile := filepath.Join(tempDir, fmt.Sprintf("error-%s.log", today))

	appBytes, err := os.ReadFile(appFile)
	if err != nil {
		t.Fatalf("expected app log file %q to exist: %v", appFile, err)
	}
	appContent := string(appBytes)
	if !strings.Contains(appContent, "sample info event") {
		t.Errorf("expected app log to contain info event, got:\n%s", appContent)
	}
	if !strings.Contains(appContent, "sample error event") {
		t.Errorf("expected app log to contain error event, got:\n%s", appContent)
	}

	errBytes, err := os.ReadFile(errFile)
	if err != nil {
		t.Fatalf("expected error log file %q to exist: %v", errFile, err)
	}
	errContent := string(errBytes)
	if strings.Contains(errContent, "sample info event") {
		t.Errorf("expected error log NOT to contain info event, got:\n%s", errContent)
	}
	if !strings.Contains(errContent, "sample error event") {
		t.Errorf("expected error log to contain error event, got:\n%s", errContent)
	}
}
