package game

import (
	"context"
	"fmt"
	"github.com/brayanMuniz/mondainai/internal/llm"
)

type Game struct {
	SessionId                  string
	Score                      int
	CharacterInfo              llm.Character
	AllowedTime                int // in seconds, client side will count it down
	CurrentYen                 int
	JLPTLevel                  JLPTLevel
	ModelSession               llm.ChatSession
	MessageHistory             llm.MessageHistory
	UserRecalledTheseFacts     map[string]bool
	CurrentCharacterEmotion    llm.Emotion
	HappyScore                 int
	CurrentRecallHint          string
	IsCurrentRecallOppurtunity bool
}

func NewGameSession(sessionId string, characterInfo llm.Character, modelSession llm.ChatSession, jlptLevel JLPTLevel) (*Game, error) {

	if sessionId == "" {
		return nil, fmt.Errorf("no session id provided")
	}
	if characterInfo.Name == "" {
		return nil, fmt.Errorf("no character info provided")
	}
	if modelSession == nil {
		return nil, fmt.Errorf("no modelSession provided")
	}

	if jlptLevel < N1 || jlptLevel > N5 {
		return nil, fmt.Errorf("invalid jlpt level")
	}

	messageHistory, err := modelSession.GetMessageHistory(context.Background())
	if err != nil {
		return nil, err
	}

	// NOTE: in the future, CurrentYen and JLPTLevel will be customize based on JLPT level
	return &Game{
		SessionId:                  sessionId,
		CharacterInfo:              characterInfo,
		Score:                      0,
		AllowedTime:                30,
		CurrentYen:                 1000,
		JLPTLevel:                  jlptLevel,
		ModelSession:               modelSession,
		MessageHistory:             *messageHistory,
		UserRecalledTheseFacts:     make(map[string]bool),
		CurrentCharacterEmotion:    llm.Neutral,
		HappyScore:                 int(llm.MaxHappyScore) - int(llm.MinHappyScore)/2,
		CurrentRecallHint:          "",
		IsCurrentRecallOppurtunity: false,
	}, nil

}

func (g *Game) ProcessUserMessage(ctx context.Context, userMessage string) error {
	llmReponse, err := g.ModelSession.SendMessage(ctx, userMessage)
	if err != nil {
		return err
	}

	updatedMessageHistory, err := g.ModelSession.GetMessageHistory(ctx)
	if err != nil {
		return err
	}
	g.MessageHistory = *updatedMessageHistory

	// user recalled fact, will be awarded
	for _, recallFact := range llmReponse.Recalled {
		if _, found := g.UserRecalledTheseFacts[recallFact]; !found {
			g.UserRecalledTheseFacts[recallFact] = true
			g.Score += 20
			g.CurrentYen += 1000
		}
	}

	// game state update based on response
	g.CurrentCharacterEmotion = llmReponse.EmotionPicture
	g.HappyScore = llmReponse.HappyScore
	g.IsCurrentRecallOppurtunity = llmReponse.RecallOppurtunity
	g.CurrentRecallHint = llmReponse.RecallHint

	return nil
}

type JLPTLevel int64

const (
	N1 JLPTLevel = iota
	N2
	N3
	N4
	N5
)

func (j JLPTLevel) String() string {
	switch j {
	case N1:
		return "N1"
	case N2:
		return "N2"
	case N3:
		return "N3"
	case N4:
		return "N4"
	case N5:
		return "N5"
	default:
		return "N5"
	}
}
