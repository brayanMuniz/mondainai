package main

import (
	// "context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// ctx := context.Background()
	// genClient, err := genai.NewClient(ctx, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//

	e := echo.New()
	e.GET("/", rootRoute)
	e.Logger.Fatal((e.Start(":1323")))
}

func setupServer() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	setupRoutes(e)

	return e
}

func setupRoutes(e *echo.Echo) {
	api := e.Group("/api")

	api.POST("/character/build", buildCharacter)
}

func buildCharacter(c echo.Context) error {
	return nil
}

func rootRoute(c echo.Context) error {
	return c.String(http.StatusOK, "こんにちわ")
}
