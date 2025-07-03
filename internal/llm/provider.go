package llm

import "context"

type Provider interface {
	BuildCharacter(ctx context.Context, name string, scenario string, characterGuide string) (*Character, error)
	NewChatSession(ctx context.Context, systemPrompt string, recallList []string) (ChatSession, error)
}

type Character struct {
	Name               string   `json:"name"`
	PersonaDescription string   `json:"persona_description"`
	Scenario           string   `json:"scenario"`
	PersonalityTraits  string   `json:"personality_traits"`
	RecallableFacts    []string `json:"recallable_facts"`
}

type LLMResponse struct {
	Message           string   `json:"message"`
	EmotionPicture    string   `json:"emotion_picture"`
	HappyScore        int      `json:"happy_score"`
	Recalled          []string `json:"recalled"`
	RecallOppurtunity bool     `json:"recall_oppurtunity"`
	RecallHint        string   `json:"recall_hint"`
}

type ChatSession interface {
	SendMessage(ctx context.Context, userMessage string) (*LLMResponse, error)
}
