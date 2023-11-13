package summary

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

type OpenAISummarizer struct {
	client *openai.Client
	prompt string  //с помощью чего gpt будет генерировать Summari
	enabled  bool
	mu sync.Mutex
}

func NewOpenAISummarizer(apiKey string, prompt string) *OpenAISummarizer {
	s:= &OpenAISummarizer {
		client: openai.NewClient(apiKey),
		prompt: prompt,
	}

	log.Printf("openai summarizer is enabled: %v", apiKey != "")

	if apiKey != "" {
		s.enabled = true //Summarizer включен
	}

	return s
}

func (s *OpenAISummarizer) Summarize(ctx context.Context, text string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.enabled {
		return "", nil
	}

	//запрос к openai
	request := openai.ChatCompletionRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: s.prompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: text,
			},
		},
		MaxTokens:   1024,
		Temperature: 1,
	}

	resp, err := s.client.CreateChatCompletion(ctx, request)

	if err != nil {
		return "", nil
	}

	rawSummary := strings.TrimSpace(resp.Choices[0].Message.Content) //выбираем 1-ый варинт
	//отрежем последнее предложение, если оно не заканчивается точкой
	if strings.HasSuffix(rawSummary, "." ) {
		return rawSummary, nil
	}
	// cut all after the last ".":
	sentences := strings.Split(rawSummary, ".")

	return strings.Join(sentences[:len(sentences)-1], ".") + ".", nil

}