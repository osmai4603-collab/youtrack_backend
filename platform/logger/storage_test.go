package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestLogStorageResultSummarizesImportantFieldsAndIncludesOperation(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	ctx := WithContext(context.Background(), logger)

	LogStorageResult(ctx, "users.FindByID", map[string]any{
		"user_id":             "user-1",
		"state":               "active",
		"failed_login_count":  2,
		"email":               "alice@example.com",
		"password_hash":       "password-value",
		"session_id":          "session-value",
		"unimportant_details": "large payload that should not be logged",
		"roles": []any{
			map[string]any{"id": "role-1", "name": "Admin"},
			map[string]any{"id": "role-2", "name": "Member"},
		},
		"swimlanes": []any{
			map[string]any{
				"subsystem": map[string]any{"id": "sub-1", "name": "Private subsystem name"},
				"columns": []any{
					map[string]any{
						"state": "open",
						"cards": []any{map[string]any{
							"id": "issue-1", "issueKey": "APP-1", "summary": "Private issue summary",
						}},
					},
				},
			},
		},
	})

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("unmarshal log record: %v", err)
	}
	if record["op"] != "users.FindByID" {
		t.Fatalf("op = %v, want users.FindByID", record["op"])
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(record["result"].(string)), &result); err != nil {
		t.Fatalf("unmarshal logged result: %v", err)
	}
	if result["user_id"] != "user-1" || result["state"] != "active" || result["failed_login_count"] != float64(2) {
		t.Errorf("important fields missing from summary: %v", result)
	}
	for _, excluded := range []string{"email", "password_hash", "session_id", "unimportant_details"} {
		if _, exists := result[excluded]; exists {
			t.Errorf("unimportant or sensitive field %q was logged: %v", excluded, result[excluded])
		}
	}

	roles := result["roles"].(map[string]any)
	if roles["count"] != float64(2) {
		t.Errorf("roles count = %v, want 2", roles["count"])
	}
	samples := roles["samples"].([]any)
	if len(samples) != 2 || samples[0].(map[string]any)["id"] != "role-1" {
		t.Errorf("unexpected role samples: %v", samples)
	}
	if _, exists := samples[0].(map[string]any)["name"]; exists {
		t.Errorf("role name should not be included in summary: %v", samples[0])
	}

	swimlanes := result["swimlanes"].(map[string]any)
	lane := swimlanes["samples"].([]any)[0].(map[string]any)
	subsystem := lane["subsystem"].(map[string]any)
	if subsystem["id"] != "sub-1" {
		t.Errorf("subsystem ID missing from board summary: %v", subsystem)
	}
	columns := lane["columns"].(map[string]any)["samples"].([]any)
	column := columns[0].(map[string]any)
	cards := column["cards"].(map[string]any)["samples"].([]any)
	card := cards[0].(map[string]any)
	if card["id"] != "issue-1" || card["state"] != "open" {
		t.Errorf("important board card fields missing: %v", card)
	}
	if _, exists := card["summary"]; exists {
		t.Errorf("issue summary should not be logged: %v", card)
	}
}

func TestLogStorageResultLimitsListSamples(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	ctx := WithContext(context.Background(), logger)

	items := make([]any, 8)
	for i := range items {
		items[i] = map[string]any{"id": i + 1, "state": "open"}
	}
	LogStorageResult(ctx, "issues.ListVisibleToUser", items)

	var record map[string]any
	if err := json.Unmarshal(output.Bytes(), &record); err != nil {
		t.Fatalf("unmarshal log record: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(record["result"].(string)), &result); err != nil {
		t.Fatalf("unmarshal logged result: %v", err)
	}
	if result["count"] != float64(8) {
		t.Errorf("count = %v, want 8", result["count"])
	}
	if got := len(result["samples"].([]any)); got != storageLogSampleLimit {
		t.Errorf("sample count = %d, want %d", got, storageLogSampleLimit)
	}
}
