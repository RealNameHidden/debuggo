package gpt

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type GPTClient struct {
	apiKey string
	client *openai.Client
}

func NewGPTClient(apiKey string) *GPTClient {
	client := openai.NewClient(apiKey)
	return &GPTClient{apiKey: apiKey, client: client}
}

func (g *GPTClient) GenerateFix(userInput string, similarDocs []string) (string, error) {
	ctx := context.Background()

	// Step 1: Build the prompt
	var builder strings.Builder

	builder.WriteString("You are an infrastructure assistant. A user has pasted a log/config snippet and you must help debug it.\n\n")
	builder.WriteString("New issue:\n")
	builder.WriteString(userInput)
	builder.WriteString("\n\n")

	if len(similarDocs) > 0 {
		builder.WriteString("Similar past issues:\n")
		for i, doc := range similarDocs {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, doc))
		}
	}

	builder.WriteString("\n---\nPlease respond with:\n- Likely Root Cause\n- Suggested Fix\n")

	// Step 2: Prepare messages for GPT
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are a senior DevOps engineer. Help troubleshoot infra issues like logs, Kubernetes configs, and Terraform plans.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: builder.String(),
		},
	}

	resp, err := g.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4, // or GPT3Turbo for cheaper
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from GPT")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}
