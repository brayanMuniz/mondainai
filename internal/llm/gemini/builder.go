package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/brayanMuniz/mondainai/internal/llm"
	"google.golang.org/genai"
)

type GeminiCharacterBuilder struct {
	baseClient *genai.Client
}

func NewGeminiCharBuilder(client *genai.Client) *GeminiCharacterBuilder {
	return &GeminiCharacterBuilder{
		baseClient: client,
	}
}

func (g *GeminiCharacterBuilder) BuildCharacter(ctx context.Context, genaiClient *genai.Client,
	name string, scenario string, characterGuide string) (*llm.Character, error) {

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

	res, err := genaiClient.Models.GenerateContent(
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
