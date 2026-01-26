package agenttests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"career-koala/agents"
	ckdb "career-koala/db"
)

func TestMaybeCaptureWriteStoresPending(t *testing.T) {
	ctx := context.Background()
	reply := "```json\n{\"write_requests\":[{\"action\":\"insert\",\"table\":\"job_applications\",\"records\":[{\"job_title\":\"Engineer\"}]}]}\n```"

	prompt, ok := agents.MaybeCaptureWrite("u1", "s1", []string{reply})
	if !ok {
		t.Fatalf("expected write capture")
	}
	if !strings.Contains(prompt, "insert 1 -> job_applications") {
		t.Fatalf("unexpected prompt: %q", prompt)
	}

	followup, handled := agents.HandlePendingWrite(ctx, "u1", "s1", "maybe", nil)
	if !handled {
		t.Fatalf("expected pending write to be handled")
	}
	if !strings.Contains(followup, "Pending write request detected") {
		t.Fatalf("unexpected followup: %q", followup)
	}
	agents.HandlePendingWrite(ctx, "u1", "s1", "no", nil)
}

func TestHandlePendingWriteDecline(t *testing.T) {
	ctx := context.Background()
	reply := "```json\n{\"write_requests\":[{\"action\":\"insert\",\"table\":\"job_applications\",\"records\":[{\"job_title\":\"Engineer\"}]}]}\n```"
	if _, ok := agents.MaybeCaptureWrite("u2", "s2", []string{reply}); !ok {
		t.Fatalf("expected write capture")
	}

	reply, handled := agents.HandlePendingWrite(ctx, "u2", "s2", "no", nil)
	if !handled {
		t.Fatalf("expected pending write handler to handle")
	}
	if !strings.Contains(reply, "skipping the write") {
		t.Fatalf("unexpected reply: %q", reply)
	}

	_, handled = agents.HandlePendingWrite(ctx, "u2", "s2", "no", nil)
	if handled {
		t.Fatalf("expected pending write to be cleared")
	}
}

func TestHandlePendingWritePrompt(t *testing.T) {
	ctx := context.Background()
	reply := "```json\n{\"write_requests\":[{\"action\":\"insert\",\"table\":\"job_applications\",\"records\":[{\"job_title\":\"Engineer\"}]}]}\n```"
	if _, ok := agents.MaybeCaptureWrite("u3", "s3", []string{reply}); !ok {
		t.Fatalf("expected write capture")
	}

	reply, handled := agents.HandlePendingWrite(ctx, "u3", "s3", "maybe", nil)
	if !handled {
		t.Fatalf("expected pending write handler to handle")
	}
	if !strings.Contains(reply, "Pending write request detected") {
		t.Fatalf("unexpected reply: %q", reply)
	}
	agents.HandlePendingWrite(ctx, "u3", "s3", "no", nil)
}

func TestHandlePendingWriteApplyWithNilDB(t *testing.T) {
	ctx := context.Background()
	reply := "```json\n{\"write_requests\":[{\"action\":\"insert\",\"table\":\"job_applications\",\"records\":[{\"job_title\":\"Engineer\"}]}]}\n```"
	if _, ok := agents.MaybeCaptureWrite("u4", "s4", []string{reply}); !ok {
		t.Fatalf("expected write capture")
	}

	reply, handled := agents.HandlePendingWrite(ctx, "u4", "s4", "yes", nil)
	if !handled {
		t.Fatalf("expected pending write handler to handle")
	}
	if !strings.Contains(reply, "database unavailable") {
		t.Fatalf("unexpected reply: %q", reply)
	}

	_, handled = agents.HandlePendingWrite(ctx, "u4", "s4", "yes", nil)
	if handled {
		t.Fatalf("expected pending write to be cleared")
	}
}

func TestWritePayloadDecodeJSONNumber(t *testing.T) {
	raw := "```json\n{\"write_requests\":[{\"action\":\"insert\",\"table\":\"coding_problems\",\"records\":[{\"leetcode_number\":1}]}]}\n```"
	payload, _, _, err := ckdb.ExtractWritePayload(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if payload == nil {
		t.Fatalf("expected payload")
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if !strings.Contains(string(data), "coding_problems") {
		t.Fatalf("unexpected payload: %s", string(data))
	}
}
