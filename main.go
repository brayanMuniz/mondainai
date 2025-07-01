package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/brayanMuniz/mondainai/internal/llm"
	"github.com/brayanMuniz/mondainai/internal/llm/gemini"
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
	genClient, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	userCharName := "ぼっちちゃん"
	userScenario := "放課後の教室"
	userCharacterGuide := "とても内気で、緊張するとすぐに挙動不審になる。ギターが命で、陰でこっそり練習している。でも、実は友達が欲しいと思っている。"

	charBuilderClient := gemini.NewGeminiCharBuilder(genClient)
	charData, err := charBuilderClient.BuildCharacter(ctx, genClient, userCharName, userScenario, userCharacterGuide)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Character Builder ---")
	fmt.Println("name ", charData.Name)
	fmt.Println("scenario ", charData.Scenario)
	fmt.Println("recallable facts ", charData.RecallableFacts)
	fmt.Println("persona description ", charData.PersonaDescription)
	fmt.Println("personality traits ", charData.PersonalityTraits)
	fmt.Println("---")

	gameplaySystemPrompt := fmt.Sprintf(`あなたは「%s」です。
これは「%s」というシナリオです。

%s

ユーザーとの会話は日本語（N4レベル）で行ってください。自然で、キャラクターの性格に合った話し方をしてください。
日本語の単語や文法を説明したり、英語に翻訳したり、ロム字を使ったりしないでください。
フリガナ（例：漢字[かんじ] や 漢字(かんじ) の形式）をあなたの応答に含めないでください。フリガナは別のシステムが追加します。
会話の中で、ユーザーが過去に言ったことやあなたのプロフィール情報（趣味、好きなもの、嫌いなものなど）について言及できるような質問やリアクションをしてください。
`,
		charData.Name,
		charData.Scenario,
		charData.PersonaDescription,
	)

	var chatProvider llm.LLMChatProvider
	chatProvider, err = gemini.NewGeminiChatClient(ctx, genClient, gameplaySystemPrompt, charData.RecallableFacts)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		text := scanner.Text()

		if len(text) != 0 {
			resp, err := chatProvider.SendMessage(ctx, text)
			if err != nil {
				log.Println(err)
				break
			}
			log.Println(resp)

		} else {
			break
		}
	}

}
