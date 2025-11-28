package openAI

import (
	"context"
	"github.com/openai/openai-go/v3"
	opt "github.com/openai/openai-go/v3/option"
)

type Manager struct {
	client openai.Client
}

func NewManager(token string) *Manager {
	return &Manager{client: openai.NewClient(
		opt.WithAPIKey(token),
	)}
}

func (s *Manager) TestRequest() {
	chatCompletion, err := s.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("Say this is a test"),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		panic(err.Error())
	}
	println(chatCompletion.Choices[0].Message.Content)
}

func (s *Manager) AnswerReview(ctx context.Context, request string) (string, error) {
	model := "ft:gpt-4.1-mini-2025-04-14:vommy-team:reviews:CfY3KAFG"

	chatCompletion, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(request),
		},
		Model: model,
	})
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, nil
}
