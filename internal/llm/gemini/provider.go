package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brayanMuniz/mondainai/internal/llm"
	"google.golang.org/genai"
)

type Provider struct {
	client *genai.Client
}

type ChatSession struct {
	chat *genai.Chat
}

func NewProvider(ctx context.Context, apiKey string) (*Provider, error) {
	genClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &Provider{
		client: genClient,
	}, nil
}

func (p *Provider) BuildCharacter(ctx context.Context, name string, scenario string, characterGuide string) (*llm.Character, error) {

	systemInstruction := `あなたはキャラクター生成アシスタントです。ユーザーの要望に基づき、アニメ風のキャラクターのプロフィール情報を生成してください。
キャラクターの具体的な性格描写（ゲームプレイLLM用のプロンプトの一部として機能します）、人間が理解しやすい性格特性の要約、そしてゲーム内でユーザーが言及するとポイントになる「思い出せる事実」のリストを含めてください。`

	config := &genai.GenerateContentConfig{
		ResponseMIMEType:  "application/json",
		SystemInstruction: genai.NewContentFromText(systemInstruction, genai.RoleUser),
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"persona_description": {Type: genai.TypeString},
				"personality_traits":  {Type: genai.TypeString},
				"recallable_facts": {
					Type: genai.TypeArray,
					Items: &genai.Schema{
						Type: genai.TypeString,
					},
				},
			},
			PropertyOrdering: []string{"persona_description", "personality_traits", "recallable_facts"},
		},
	}

	buildMessage := fmt.Sprintf(`以下のキャラクター情報に基づいて、JSONを生成してください：
	キャラクター名: %s
	シナリオ: %s
	キャラクターガイド: %s`, name, scenario, characterGuide)

	res, err := p.client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(buildMessage),
		config,
	)

	if err != nil {
		return nil, err
	}

	// "Inappropriate" message sent
	if len(res.Candidates) == 0 || res.Candidates[0].Content == nil || len(res.Candidates[0].Content.Parts) == 0 {
		return nil, err
	}

	chatResponse := res.Candidates[0].Content.Parts[0].Text

	var charData llm.Character
	err = json.Unmarshal([]byte(chatResponse), &charData)
	if err != nil {
		return nil, err
	}

	charData.Name = name
	charData.Scenario = scenario

	return &charData, nil
}

func (p *Provider) NewChatSession(ctx context.Context, systemPrompt string, recallList []string) (llm.ChatSession, error) {

	var minVal float64 = 0
	var maxVal float64 = 100

	config := &genai.GenerateContentConfig{
		ResponseMIMEType:  "application/json",
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
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
	chat, err := p.client.Chats.Create(ctx, "gemini-2.5-flash", config, history)
	if err != nil {
		return nil, err
	}

	return &ChatSession{
		chat: chat,
	}, nil

}

func (s *ChatSession) SendMessage(ctx context.Context, userMessage string) (*llm.LLMResponse, error) {
	res, err := s.chat.SendMessage(ctx, genai.Part{Text: userMessage})
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

func (g *ChatSession) GetMessageHistory(ctx context.Context) (*llm.MessageHistory, error) {
	geminiHistory := g.chat.History(false)

	llmMessageHistory := &llm.MessageHistory{
		Messages: make([]llm.Message, 0),
	}

	for _, content := range geminiHistory {
		if len(content.Parts) > 0 {
			chatResponse := content.Parts[0].Text
			llmMessageHistory.Messages = append(llmMessageHistory.Messages, llm.Message{
				Text: chatResponse,
				Role: content.Role,
			})
		}
	}

	return llmMessageHistory, nil

}
