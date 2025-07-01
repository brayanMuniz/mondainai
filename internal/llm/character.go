package llm

import "context"

type LLMCharacterBuilder interface {
	BuildCharacter(ctx context.Context, userDescription string) (*Character, error)
}

type Character struct {
	Name               string   `json:"name"`
	PersonaDescription string   `json:"persona_description"`
	Scenario           string   `json:"scenario"`
	PersonalityTraits  string   `json:"personality_traits"`
	RecallableFacts    []string `json:"recallable_facts"`
}
