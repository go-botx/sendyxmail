package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"sendymail/apiv0"
	"sendymail/mutemanager"
	"sendymail/tokenmanager"
	"strings"
	"syscall"
	"time"

	"github.com/go-botx/bot"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const groupChatMailSuffix = "@chat-id.internal"

var (
	tm *tokenmanager.TokenManager
	mm *mutemanager.MuteManager
)

func main() {
	var err error
	botCreds := getEnvVarOrPanic("BOT_CREDENTIALS", 71, "BOT_CREDENTIALS must be provided as env variable in format 'cts_server@bot_secret@bot_id'")
	metadataSecret := getEnvVarOrPanic("METADATA_SECRET", 20, "METADATA_SECRET must be provided as env variable and must be at least 20 characters")
	tokenFile := getEnvVarOrPanic("TOKEN_FILE", 2, "TOKEN_FILE path must be provided as env variable")
	muteFile := getEnvVarOrPanic("MUTE_FILE", 2, "MUTE_FILE path must be provided as env variable")

	tm, err = tokenmanager.Run(tokenFile, time.Duration(10*time.Minute))
	if err != nil {
		panic(err)
	}

	mm, err = mutemanager.New(muteFile)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			os.WriteFile(muteFile, []byte{}, 0666)
			mm, err = mutemanager.New(muteFile)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}

	// This is superApp.
	// Bot subApp is mounted to /botapi
	// Service subApp is mounted to /api/v0 subApp
	app := fiber.New(fiber.Config{
		AppName:                      AppName(),
		ServerHeader:                 AppName() + " " + AppVersion(),
		ReadTimeout:                  time.Minute,
		WriteTimeout:                 time.Minute,
		IdleTimeout:                  time.Minute,
		ProxyHeader:                  fiber.HeaderXForwardedFor,
		DisableStartupMessage:        true,
		DisablePreParseMultipartForm: true,
	})

	var b *bot.Bot

	botStatusHandler := NewStatusHandler()
	botCommandHandler := NewCommandHandler()
	b, err = bot.New(botCreds,
		bot.WithDebugHTTPService(),
		bot.WithStatusHandler(botStatusHandler),
		bot.WithCommandHandler(botCommandHandler))
	if err != nil {
		panic(err)
	}
	app.Mount("/botapi", b.FiberApp())

	apiGroup := app.Group("/api")
	apiGroup.Use(logger.New())
	apiGroup.Use(recover.New())

	apiGroup.Mount("/v0", apiv0.New(apiv0.APIConfig{
		Bot:                      b,
		GroupChatMailSuffix:      groupChatMailSuffix,
		CheckBearerToken:         checkToken,
		MetadataEncryptionSecret: metadataSecret,
		CheckAllowedSend:         checkAllowedSend,
	}))

	go func() {
		if err := app.Listen(":8000"); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()
}

func checkToken(token string) error {
	if tm.HasToken(token) {
		return nil
	}
	return errors.New("token not registered in token manager")
}

func getEnvVarOrPanic(name string, minLen int, panicMessage string) string {
	value, ok := os.LookupEnv(name)
	value = strings.TrimSpace(value)
	if !ok || len(value) < minLen {
		panic(panicMessage)
	}
	return value
}

func checkAllowedSend(ident string) (err error) {

	if mm.GetMute(ident) {
		return errors.New("bot is muted in this chat")
	}
	return nil
}
