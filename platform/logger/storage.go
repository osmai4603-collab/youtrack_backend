package logger

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
)

const storageLogSampleLimit = 5

var importantStorageFields = map[string]struct{}{
	"activesprint":     {},
	"cards":            {},
	"archived":         {},
	"agileid":          {},
	"assigned":         {},
	"banned":           {},
	"columncounts":     {},
	"columns":          {},
	"createdat":        {},
	"deleted":          {},
	"expiresat":        {},
	"favorite":         {},
	"groupid":          {},
	"headers":          {},
	"id":               {},
	"islocked":         {},
	"isowner":          {},
	"isstarted":        {},
	"issueid":          {},
	"issuekey":         {},
	"issuetype":        {},
	"permissions":      {},
	"pinned":           {},
	"priority":         {},
	"projectid":        {},
	"purged":           {},
	"resourcetype":     {},
	"revoked":          {},
	"roleid":           {},
	"roles":            {},
	"scope":            {},
	"scopeid":          {},
	"scopetype":        {},
	"state":            {},
	"status":           {},
	"subsystem":        {},
	"subsystems":       {},
	"swimlanes":        {},
	"sprints":          {},
	"updated":          {},
	"updatedat":        {},
	"userid":           {},
	"visible":          {},
	"permissionid":     {},
	"permissionname":   {},
	"failedlogincount": {},
}

var sensitiveStorageFields = map[string]struct{}{
	"accesstoken":   {},
	"authorization": {},
	"cookie":        {},
	"data":          {},
	"email":         {},
	"phone":         {},
}

// LogStorageResult records an operation and a bounded summary of its result.
func LogStorageResult(ctx context.Context, operation string, result any) {
	logger := FromContext(ctx)
	encoded, err := json.Marshal(result)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "storage result could not be serialized",
			slog.String("op", operation),
			slog.Any("error", err),
		)
		return
	}

	var value any
	if err := json.Unmarshal(encoded, &value); err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "storage result could not be decoded for redaction",
			slog.String("op", operation),
			slog.Any("error", err),
		)
		return
	}

	encoded, err = json.Marshal(summarizeStorageValue(redactStorageValue(value)))
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "storage result could not be serialized after redaction",
			slog.String("op", operation),
			slog.Any("error", err),
		)
		return
	}

	if len(encoded) != 0 {
		logger.LogAttrs(ctx, slog.LevelDebug, operation,
			slog.String("result", string(encoded)),
		)
	}

}

func summarizeStorageValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		summary := make(map[string]any)
		for key, child := range value {
			field := normalizeStorageField(key)
			if _, important := importantStorageFields[field]; important {
				summary[key] = summarizeStorageValue(child)
			}
		}
		return summary
	case []any:
		samples := make([]any, 0, min(len(value), storageLogSampleLimit))
		for _, item := range value[:min(len(value), storageLogSampleLimit)] {
			samples = append(samples, summarizeStorageValue(item))
		}
		return map[string]any{
			"count":   len(value),
			"samples": samples,
		}
	default:
		return value
	}
}

func redactStorageValue(value any) any {
	switch value := value.(type) {
	case map[string]any:
		for key, child := range value {
			field := normalizeStorageField(key)
			_, sensitive := sensitiveStorageFields[field]
			sensitive = sensitive || strings.Contains(field, "password") ||
				strings.Contains(field, "token") ||
				strings.Contains(field, "secret") ||
				strings.Contains(field, "credential") ||
				strings.Contains(field, "apikey") ||
				strings.Contains(field, "privatekey") ||
				strings.Contains(field, "session")
			if sensitive {
				value[key] = "***"
			} else {
				value[key] = redactStorageValue(child)
			}
		}
		return value
	case []any:
		for i, child := range value {
			value[i] = redactStorageValue(child)
		}
		return value
	default:
		return value
	}
}

func normalizeStorageField(field string) string {
	field = strings.ToLower(strings.TrimSpace(field))
	field = strings.ReplaceAll(field, "_", "")
	field = strings.ReplaceAll(field, "-", "")
	return field
}
