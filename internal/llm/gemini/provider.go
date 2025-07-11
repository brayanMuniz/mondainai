package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/brayanMuniz/mondainai/internal/llm"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	client *genai.Client
}

type GeminiChatSession struct {
	chat *genai.Chat
}

func NewProvider(ctx context.Context, apiKey string) (*GeminiProvider, error) {
	genClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &GeminiProvider{
		client: genClient,
	}, nil
}

func (p *GeminiProvider) BuildCharacter(ctx context.Context, name string, scenario string, characterGuide string) (*llm.Character, error) {
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

func (p *GeminiProvider) NewChatSession(ctx context.Context, systemPrompt string, recallList []string) (llm.ChatSession, error) {
	config := newGameplayConfig(systemPrompt, recallList)
	history := []*genai.Content{}
	chat, err := p.client.Chats.Create(ctx, "gemini-2.5-flash", config, history)
	if err != nil {
		return nil, err
	}

	return &GeminiChatSession{
		chat: chat,
	}, nil

}

func (p *GeminiProvider) NewSessionFromMessages(ctx context.Context, systemPrompt string, recallList []string, pastMessages llm.MessageHistory) (llm.ChatSession, error) {
	config := newGameplayConfig(systemPrompt, recallList)
	history := make([]*genai.Content, 0, len(pastMessages.Messages))
	for _, message := range pastMessages.Messages {
		content := &genai.Content{}
		if message.Role == "user" && message.UserText != nil {
			content.Role = genai.RoleUser
			content.Parts = []*genai.Part{
				{Text: *message.UserText},
			}
		} else {
			content.Role = genai.RoleModel
			content.Parts = []*genai.Part{
				{Text: message.ModelResponse.Message},
			}
		}

		history = append(history, content)
	}

	chat, err := p.client.Chats.Create(ctx, "gemini-2.5-flash", config, history)
	if err != nil {
		return nil, err
	}

	return &GeminiChatSession{
		chat: chat,
	}, nil

}

func (s *GeminiChatSession) SendMessage(ctx context.Context, userMessage string) (*llm.LLMResponse, error) {
	log.Println("LLM is sending a message")
	res, err := s.chat.SendMessage(ctx, genai.Part{Text: userMessage})
	if err != nil {
		return nil, err
	}

	if len(res.Candidates) == 0 ||
		res.Candidates[0].Content == nil ||
		len(res.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("received an empty or blocked response from Gemini API")
	}

	chatResponse := res.Candidates[0].Content.Parts[0].Text

	var formattedResponse llm.LLMResponse
	err = json.Unmarshal([]byte(chatResponse), &formattedResponse)
	if err != nil {
		return nil, err
	}

	return &formattedResponse, nil
}

func (g *GeminiChatSession) GetMessageHistory(ctx context.Context) (*llm.MessageHistory, error) {
	geminiHistory := g.chat.History(false)

	llmMessageHistory := &llm.MessageHistory{
		Messages: make([]llm.Message, 0),
	}

	for _, content := range geminiHistory {
		llmMessage := llm.Message{
			Role: content.Role,
		}

		if len(content.Parts) > 0 {
			if content.Role == "user" {
				userResponse := content.Parts[0].Text
				llmMessage.UserText = &userResponse
			} else {
				modelResponse := content.Parts[0].Text
				var responseFromModel llm.LLMResponse
				err := json.Unmarshal([]byte(modelResponse), &responseFromModel)
				if err == nil {
					llmMessage.ModelResponse = &responseFromModel
				} else {
					responseFromModel.Message = modelResponse
					llmMessage.ModelResponse = &responseFromModel
				}

			}

			llmMessageHistory.Messages = append(llmMessageHistory.Messages, llmMessage)

		}
	}

	return llmMessageHistory, nil
}

func newGameplayConfig(systemPrompt string, recallList []string) *genai.GenerateContentConfig {
	return &genai.GenerateContentConfig{
		ResponseMIMEType:  "application/json",
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"message": {Type: genai.TypeString},
				"emotion_picture": {
					Type: genai.TypeString,
					Enum: llm.GetEmotionList(),
				},
				"happy_score": {
					Type:    genai.TypeInteger,
					Minimum: &llm.MinHappyScore,
					Maximum: &llm.MaxHappyScore,
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

}
