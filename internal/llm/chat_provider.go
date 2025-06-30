package llm

import "context"

type LLMChatProvider interface {
	SendMessage(ctx context.Context, userMessage string) (*LLMResponse, error)
}

type LLMResponse struct {
	Message           string   `json:"message"`
	EmotionPicture    string   `json:"emotion_picture"`
	HappyScore        int      `json:"happy_score"`
	Recalled          []string `json:"recalled"`
	RecallOppurtunity bool     `json:"recall_oppurtunity"`
	RecallHint        string   `json:"recall_hint"`
}
