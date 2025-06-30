package gemini

import (
	"context"
	"encoding/json"

	"github.com/brayanMuniz/mondainai/internal/llm"
	"google.golang.org/genai"
)

type GeminiChatClient struct {
	client *genai.Client
	chat   *genai.Chat
}

func NewGeminiChatClient(ctx context.Context, genaiClient *genai.Client, systemInstruction string, recallList []string) (*GeminiChatClient, error) {
	var minVal float64 = 0
	var maxVal float64 = 100

	// Sets persona and scenario
	// Defines Core Personality Traits.
	// Sets Language, Naturalness, and Difficulty Level.
	// Omit Furigana
	// If the user recalls things about the character, return an indicator that the user has done so
	// recall oppurtunities: enum list
	config := &genai.GenerateContentConfig{
		ResponseMIMEType:  "application/json",
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"message": {Type: genai.TypeString},
				"emotion_picture": {
					Type: genai.TypeString,
					Enum: []string{"happy", "neutral", "sad", "embarrased", "shy"},
				},
				"happy_score": {
					Type:    genai.TypeInteger,
					Minimum: &minVal,
					Maximum: &maxVal,
				},
				"recalled": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeString,
						Enum: recallList,
					},
				},
				"recall_oppurtunity": {Type: genai.TypeBoolean},
				"recall_hint":        {Type: genai.TypeString},
			},
			PropertyOrdering: []string{"message", "emotion_picture", "happy_score", "recall_oppurtunity", "recall_hint"},
		},
	}

	history := []*genai.Content{}
	chat, err := genaiClient.Chats.Create(ctx, "gemini-2.5-flash", config, history)
	if err != nil {
		return nil, err
	}

	return &GeminiChatClient{
		client: genaiClient,
		chat:   chat,
	}, nil

}

// Use the paramters in the struct to send the message
func (g *GeminiChatClient) SendMessage(ctx context.Context, userMessage string) (*llm.LLMResponse, error) {
	res, err := g.chat.SendMessage(ctx, genai.Part{Text: userMessage})
	if err != nil {
		return nil, err
	}

	// "Inappropriate" message sent
	if len(res.Candidates) == 0 || res.Candidates[0].Content == nil || len(res.Candidates[0].Content.Parts) == 0 {
		return nil, err
	}

	chatResponse := res.Candidates[0].Content.Parts[0].Text

	var formattedResponse llm.LLMResponse
	err = json.Unmarshal([]byte(chatResponse), &formattedResponse)
	if err != nil {
		return nil, err
	}

	return &formattedResponse, nil
}
