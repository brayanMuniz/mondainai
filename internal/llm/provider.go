package llm

import "context"

var MinHappyScore float64 = 0
var MaxHappyScore float64 = 100

type Provider interface {
	BuildCharacter(ctx context.Context, name string, scenario string, characterGuide string) (*Character, error)
	NewChatSession(ctx context.Context, systemPrompt string, recallList []string) (ChatSession, error)
	NewSessionFromMessages(ctx context.Context, systemPrompt string, recallList []string, pastMessages MessageHistory) (ChatSession, error)
}

type Character struct {
	Name               string   `json:"name"`
	PersonaDescription string   `json:"persona_description"`
	Scenario           string   `json:"scenario"`
	PersonalityTraits  string   `json:"personality_traits"`
	RecallableFacts    []string `json:"recallable_facts"`
}

type ChatSession interface {
	GetMessageHistory(ctx context.Context) (*MessageHistory, error)
	SendMessage(ctx context.Context, userMessage string) (*LLMResponse, error)
}

type LLMResponse struct {
	Message           string   `json:"message"`
	EmotionPicture    Emotion  `json:"emotion_picture"` // maps to Emotion "enum"
	HappyScore        int      `json:"happy_score"`
	Recalled          []string `json:"recalled"`
	RecallOppurtunity bool     `json:"recall_oppurtunity"`
	RecallHint        string   `json:"recall_hint"`
}

type MessageHistory struct {
	Messages []Message `json:"messages"`
}

type Message struct {
	UserText      *string      `json:"userText,omitempty"`
	ModelResponse *LLMResponse `json:"modelResponse,omitempty"`
	Role          string       `json:"role"` // user or model
}

type Emotion string

const (
	Happy       Emotion = "Happy"
	Neutral     Emotion = "Neutral"
	Sad         Emotion = "Sad"
	Embarrassed Emotion = "Embarrassed"
	Shy         Emotion = "Shy"
)

func GetEmotionList() []string {
	return []string{
		string(Happy),
		string(Neutral),
		string(Sad),
		string(Embarrassed),
		string(Shy),
	}
}
