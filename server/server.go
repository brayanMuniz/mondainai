package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/brayanMuniz/mondainai/internal/llm"
	"github.com/brayanMuniz/mondainai/internal/llm/gemini"
	"github.com/brayanMuniz/mondainai/templates"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
)

var activeChatSession = make(map[string]llm.ChatSession)

type CharacterBuilderRequest struct {
	Provider       string `json:"provider"` // gemini or openai
	Name           string `json:"name"`
	Scenario       string `json:"scenario"`
	CharacterGuide string `json:"characterGuide"`
	JLPTLevel      string `json:"jlptLevel"`
}

type MessageRequest struct {
	Message string `json:"message"`
}

type Server struct {
	echo         *echo.Echo
	llmproviders map[string]llm.Provider
}

// For testing I am going to import the api key from the env
func NewServer(geminiApiKey string) (*Server, error) {
	e := echo.New()
	e.Use(middleware.Logger(), middleware.Recover(), middleware.CORS())

	providers := make(map[string]llm.Provider, 0)
	if geminiApiKey != "" {
		genProvider, err := gemini.NewProvider(context.Background(), geminiApiKey)
		if err != nil {
			return nil, err
		}
		providers["gemini"] = genProvider

	}

	// TEMP CODE ===
	tempCharFile, err := os.ReadFile("temp_char.json")
	if err != nil {
		return nil, err
	}
	var tempChar llm.Character
	err = json.Unmarshal(tempCharFile, &tempChar)
	if err != nil {
		return nil, err
	}

	gameplaySystemPrompt := fmt.Sprintf(`
		あなたは「%s」です。
		これは「%s」というシナリオです。
		%s
ユーザーとの会話は日本語（%sレベル）で行ってください。自然で、キャラクターの性格に合った話し方をしてください。
日本語の単語や文法を説明したり、英語に翻訳したり、ロム字を使ったりしないでください。
フリガナ（例：漢字[かんじ] や 漢字(かんじ) の形式）をあなたの応答に含めないでください。フリガナは別のシステムが追加します。
会話の中で、ユーザーが過去に言ったことやあなたのプロフィール情報（趣味、好きなもの、嫌いなものなど）について言及できるような質問やリアクションをしてください。
`,
		tempChar.Name,
		"N4",
		tempChar.Scenario,
		tempChar.PersonaDescription,
	)

	tempChatFile, err := os.ReadFile("temp_chat.json")
	if err != nil {
		return nil, err
	}
	var tempChatHistory llm.MessageHistory
	err = json.Unmarshal(tempChatFile, &tempChatHistory)
	if err != nil {
		return nil, err
	}

	var chatSession llm.ChatSession
	chatSession, err = providers["gemini"].NewSessionFromMessages(context.Background(), gameplaySystemPrompt, tempChar.RecallableFacts, tempChatHistory)
	if err != nil {
		return nil, err
	}

	activeChatSession["dev"] = chatSession
	// TEMP CODE ====

	s := &Server{
		echo:         e,
		llmproviders: providers,
	}

	s.setupRoutes()
	return s, nil
}

func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

func (s *Server) setupRoutes() {
	s.echo.Static("/", "views")

	api := s.echo.Group("/api")
	api.POST("/character/build", s.buildCharacter)

	api.GET("/character/chat/:sessionId", s.getChatPage)
	api.POST("/character/chat/:sessionId", s.sendMessage)
}

func (s *Server) buildCharacter(c echo.Context) error {
	req := new(CharacterBuilderRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	provider, ok := s.llmproviders[req.Provider]
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("provide a valid provider"))
	}

	ctx := c.Request().Context()
	charData, err := provider.BuildCharacter(ctx, req.Name, req.Scenario, req.CharacterGuide)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	gameplaySystemPrompt := fmt.Sprintf(`
		あなたは「%s」です。
		これは「%s」というシナリオです。
		%s
ユーザーとの会話は日本語（%sレベル）で行ってください。自然で、キャラクターの性格に合った話し方をしてください。
日本語の単語や文法を説明したり、英語に翻訳したり、ロム字を使ったりしないでください。
フリガナ（例：漢字[かんじ] や 漢字(かんじ) の形式）をあなたの応答に含めないでください。フリガナは別のシステムが追加します。
会話の中で、ユーザーが過去に言ったことやあなたのプロフィール情報（趣味、好きなもの、嫌いなものなど）について言及できるような質問やリアクションをしてください。
`,
		charData.Name,
		req.JLPTLevel,
		charData.Scenario,
		charData.PersonaDescription,
	)

	var chatSession llm.ChatSession
	chatSession, err = provider.NewChatSession(ctx, gameplaySystemPrompt, charData.RecallableFacts)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	sessionId, err := generateUniqeSessionId()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}
	log.Println("saving sessionId", sessionId)
	activeChatSession[sessionId] = chatSession

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return templates.CharacterCreated(charData.Name, req.JLPTLevel, charData.RecallableFacts, charData.PersonaDescription, charData.PersonalityTraits, sessionId).Render(c.Request().Context(), c.Response().Writer)

}

func (s *Server) sendMessage(c echo.Context) error {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("no sessionId provided"))
	}

	chatSession, ok := activeChatSession[sessionId]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("Your session was not found"))
	}

	req := new(MessageRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	llmReponse, err := chatSession.SendMessage(ctx, req.Message)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, llmReponse)
}

func (s *Server) getChatPage(c echo.Context) error {
	sessionId := c.Param("sessionId")
	if sessionId == "" {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("no sessionId provided"))
	}

	chatSession, ok := activeChatSession[sessionId]
	if !ok {
		return echo.NewHTTPError(http.StatusNotFound, fmt.Errorf("Your session was not found"))
	}

	ctx := c.Request().Context()
	messageHistory, err := chatSession.GetMessageHistory(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Could not get your message history"))
	}

	// c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	// return templates.ChatInterface(sessionId, messageHistory.Messages).Render(c.Request().Context(), c.Response().Writer)

	return c.JSON(http.StatusOK, messageHistory)

}

func generateUniqeSessionId() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
