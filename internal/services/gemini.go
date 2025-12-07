package services

import (
    "context"
    "fmt"
    "os"

    "google.golang.org/genai"
)

type GeminiService struct {
    client *genai.Client
    ctx    context.Context
}

func NewGeminiService(apiKey string) *GeminiService {
    ctx := context.Background()

    // Set API key as environment variable or pass it directly
    // The SDK reads from GEMINI_API_KEY env var by default
    if apiKey != "" {
        os.Setenv("GEMINI_API_KEY", apiKey)
    }

    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        // Handle error - maybe return nil and check in calling code
        panic(fmt.Sprintf("Failed to create Gemini client: %v", err))
    }

    return &GeminiService{
        client: client,
        ctx:    ctx,
    }
}

func (s *GeminiService) GenerateContent(prompt string) (string, error) {
    result, err := s.client.Models.GenerateContent(
        s.ctx,
        "gemini-2.5-flash",
        genai.Text(prompt),
        nil,
    )
    if err != nil {
        return "", fmt.Errorf("failed to generate content: %w", err)
    }

    return result.Text(), nil
}