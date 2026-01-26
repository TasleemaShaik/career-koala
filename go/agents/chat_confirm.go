package agents

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	ckdb "career-koala/db"
)

type pendingWrite struct {
	Payload ckdb.WritePayload
	RawJSON string
	Summary string
}

var pendingWrites = struct {
	sync.Mutex
	items map[string]pendingWrite
}{
	items: make(map[string]pendingWrite),
}

func HandlePendingWrite(ctx context.Context, userID, sessionID, message string, dbConn *sql.DB) (string, bool) {
	key := pendingKey(userID, sessionID)
	pendingWrites.Lock()
	pending, ok := pendingWrites.items[key]
	pendingWrites.Unlock()
	if !ok {
		return "", false
	}

	answer := strings.TrimSpace(strings.ToLower(message))
	if isAffirmative(answer) {
		if dbConn == nil {
			pendingWrites.Lock()
			delete(pendingWrites.items, key)
			pendingWrites.Unlock()
			return "Failed to apply writes: database unavailable", true
		}
		wctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		result, err := ckdb.ApplyWriteRequests(wctx, dbConn, pending.Payload)
		pendingWrites.Lock()
		delete(pendingWrites.items, key)
		pendingWrites.Unlock()
		if err != nil {
			return fmt.Sprintf("Failed to apply writes: %v", err), true
		}
		return fmt.Sprintf("Applied writes: %s", result), true
	}
	if isNegative(answer) {
		pendingWrites.Lock()
		delete(pendingWrites.items, key)
		pendingWrites.Unlock()
		return "Okay, skipping the write. Let me know how else I can help.", true
	}

	return fmt.Sprintf("Pending write request detected. Reply \"yes\" to apply or \"no\" to skip.\n\n%s", pending.RawJSON), true
}

func MaybeCaptureWrite(userID, sessionID string, replies []string) (string, bool) {
	for _, reply := range replies {
		payload, rawJSON, summary, err := ckdb.ExtractWritePayload(reply)
		if err != nil || payload == nil {
			continue
		}
		key := pendingKey(userID, sessionID)
		pendingWrites.Lock()
		pendingWrites.items[key] = pendingWrite{
			Payload: *payload,
			RawJSON: rawJSON,
			Summary: summary,
		}
		pendingWrites.Unlock()
		prompt := fmt.Sprintf("I can apply the following write request(s): %s\n\nReply \"yes\" to insert, or \"no\" to skip.\n\n```json\n%s\n```", summary, rawJSON)
		return prompt, true
	}
	return "", false
}

func pendingKey(userID, sessionID string) string {
	return userID + ":" + sessionID
}

func isAffirmative(msg string) bool {
	for _, token := range strings.Fields(msg) {
		switch strings.Trim(token, ".,!?;:") {
		case "yes", "y", "yeah", "yep", "sure", "confirm", "ok", "okay":
			return true
		}
	}
	return false
}

func isNegative(msg string) bool {
	for _, token := range strings.Fields(msg) {
		switch strings.Trim(token, ".,!?;:") {
		case "no", "n", "nope", "nah", "cancel":
			return true
		}
	}
	return false
}
