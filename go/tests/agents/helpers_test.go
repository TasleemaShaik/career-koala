package agenttests

import (
	"context"
	"testing"

	"iter"

	adkagent "google.golang.org/adk/agent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type fakeLLM struct {
	name  string
	reply string
}

func (f fakeLLM) Name() string {
	if f.name != "" {
		return f.name
	}
	return "fake-model"
}

func (f fakeLLM) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		text := f.reply
		if text == "" {
			text = "mock response"
		}
		resp := &model.LLMResponse{
			Content: &genai.Content{
				Role:  genai.RoleModel,
				Parts: []*genai.Part{{Text: text}},
			},
		}
		yield(resp, nil)
	}
}

func runAgent(t *testing.T, a adkagent.Agent, msg string) []string {
	t.Helper()

	ctx := context.Background()
	sessSvc := session.InMemoryService()
	rnr, err := runner.New(runner.Config{
		AppName:        "test_app",
		Agent:          a,
		SessionService: sessSvc,
	})
	if err != nil {
		t.Fatalf("runner: %v", err)
	}

	createResp, err := sessSvc.Create(ctx, &session.CreateRequest{
		AppName: "test_app",
		UserID:  "test_user",
	})
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	seq := rnr.Run(ctx, "test_user", createResp.Session.ID(), &genai.Content{
		Parts: []*genai.Part{{Text: msg}},
	}, adkagent.RunConfig{})

	var replies []string
	for event, err := range seq {
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if event == nil || event.LLMResponse.Content == nil || event.Author == "user" {
			continue
		}
		for _, part := range event.LLMResponse.Content.Parts {
			if part != nil && part.Text != "" {
				replies = append(replies, part.Text)
			}
		}
	}

	if len(replies) == 0 {
		t.Fatalf("no replies returned")
	}

	return replies
}
