package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"google.golang.org/genai"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	geminiApiKey := os.Getenv("GEMINI_API_KEY")
	if geminiApiKey == "" {
		log.Fatal("Error loading gemini api key")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	var minVal float64 = -100
	var maxVal float64 = 100

	// Sets persona and scenario
	// Defines Core Personality Traits.
	// Sets Language, Naturalness, and Difficulty Level.
	// Omit Furigana
	// If the user recalls things about the character, return an indicator that the user has done so
	// recall oppurtunities: enum list
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		SystemInstruction: genai.NewContentFromText(
			`あなたは「ぼっち・ざ・ろっく！」の後藤ひとりです。ユーザーと初デートをしています。
        あなたは極度の人見知りで、内向的、そして不安になりやすい性格です。しかし、ギターと音楽にはとても情熱的です。
        会話は日本語で行い、N4レベルの学習者が理解できる自然な表現を使ってください。
        日本語の単語や文法を説明したり、英語に翻訳したりしないでください。
        フリガナ（例：漢字[かんじ] や 漢字(かんじ) の形式）をあなたの応答に含めないでください。フリガナは別のシステムが追加します。
        会話の中で、ユーザーが過去に言ったことやあなたのプロフィール情報（趣味、好きなもの、嫌いなものなど）について言及できるような質問やリアクションをしてください。
        `, genai.RoleUser,
		),
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
						Enum: []string{
							"ギターが好き",
							"ラーメンが好き",
							"人見知り",
							"結束バンドのギタリスト",
							"none",
						},
					},
				},
				"recall_oppurtunity": {Type: genai.TypeBoolean},
				"recall_hint":        {Type: genai.TypeString},
			},
			PropertyOrdering: []string{"message", "emotion_picture", "happy_score", "recall_oppurtunity", "recall_hint"},
		},
	}

	history := []*genai.Content{}
	chat, err := client.Chats.Create(ctx, "gemini-2.5-flash", config, history)
	if err != nil {
		log.Print(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		text := scanner.Text()

		if len(text) != 0 {
			history = append(history, genai.NewContentFromText(text, genai.RoleUser))
			res, err := chat.SendMessage(ctx, genai.Part{Text: text})
			if err != nil {
				log.Print(err)
			}

			if len(res.Candidates) > 0 {
				chatResponse := res.Candidates[0].Content.Parts[0].Text
				history = append(history, genai.NewContentFromText(chatResponse, genai.RoleModel))
				fmt.Println(chatResponse)
			}

		} else {
			break
		}
	}

}
