package db

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestExtractWritePayloadFromFence(t *testing.T) {
	text := "```json\n{\"write_requests\":[{\"action\":\"insert\",\"table\":\"job_applications\",\"records\":[{\"job_title\":\"Engineer\"}]}]}\n```"
	payload, raw, summary, err := ExtractWritePayload(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil || len(payload.WriteRequests) != 1 {
		t.Fatalf("expected payload with one request")
	}
	if !strings.Contains(summary, "insert 1 -> job_applications") {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if raw == "" || !strings.Contains(raw, "job_applications") {
		t.Fatalf("unexpected raw json: %q", raw)
	}
}

func TestExtractWritePayloadRawJSON(t *testing.T) {
	text := "{\"write_requests\":[{\"action\":\"insert\",\"table\":\"projects\",\"records\":[{\"name\":\"Koala\"}]}]}"
	payload, raw, summary, err := ExtractWritePayload(text)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil || len(payload.WriteRequests) != 1 {
		t.Fatalf("expected payload with one request")
	}
	if raw == "" || !strings.Contains(raw, "projects") {
		t.Fatalf("unexpected raw json: %q", raw)
	}
	if summary == "" {
		t.Fatalf("expected summary")
	}
}

func TestExtractWritePayloadInvalidAction(t *testing.T) {
	text := "```json\n{\"write_requests\":[{\"action\":\"update\",\"table\":\"projects\",\"records\":[{\"name\":\"Koala\"}]}]}\n```"
	payload, _, _, err := ExtractWritePayload(text)
	if err == nil {
		t.Fatalf("expected error for invalid action, got nil payload=%v", payload)
	}
}

func TestExtractWritePayloadEmpty(t *testing.T) {
	payload, raw, summary, err := ExtractWritePayload("no json here")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload != nil || raw != "" || summary != "" {
		t.Fatalf("expected empty result, got payload=%v raw=%q summary=%q", payload, raw, summary)
	}
}

func TestValueHelpers(t *testing.T) {
	rec := map[string]interface{}{
		"title":          "Engineer",
		"already_solved": json.Number("1"),
		"leetcode":       json.Number("42"),
		"contact_id":     json.Number("7"),
		"start_date":     "2025-12-01",
		"stack":          []interface{}{"go", "sql"},
	}

	if got := getString(rec, "title"); got != "Engineer" {
		t.Fatalf("getString: %q", got)
	}
	if got := getBool(rec, "already_solved"); got != true {
		t.Fatalf("getBool: %v", got)
	}
	if got := getInt(rec, "leetcode"); got != 42 {
		t.Fatalf("getInt: %d", got)
	}
	if got := getIntPtr(rec, "contact_id"); got == nil || *got != 7 {
		t.Fatalf("getIntPtr: %v", got)
	}
	if got := getStringSlice(rec, "stack"); len(got) != 2 || got[0] != "go" {
		t.Fatalf("getStringSlice: %v", got)
	}
	date := getDate(rec, "start_date")
	if date == nil {
		t.Fatalf("getDate: nil")
	}
	want := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	if !date.Equal(want) {
		t.Fatalf("getDate: %v", date)
	}
}
