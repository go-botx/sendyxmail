package apiv0

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/go-botx/bot"
	"github.com/go-botx/bot/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type CheckBearerTokenFunc func(token string) error
type CheckAllowedSendFunc func(ident string) error

type APIConfig struct {
	Bot                      *bot.Bot
	GroupChatMailSuffix      string
	CheckBearerToken         CheckBearerTokenFunc
	MetadataEncryptionSecret string
	CheckAllowedSend         CheckAllowedSendFunc
}

var apiCtxConfigKey = uuid.MustParse("a30f42ca-d68a-4229-b868-add3792f512a") // This is random UUID

func New(config APIConfig) *fiber.App {
	apiConfig := &APIConfig{
		Bot:                      config.Bot,
		GroupChatMailSuffix:      config.GroupChatMailSuffix,
		CheckBearerToken:         config.CheckBearerToken,
		MetadataEncryptionSecret: config.MetadataEncryptionSecret,
		CheckAllowedSend:         config.CheckAllowedSend,
	}
	api := fiber.New()
	api.Use(injectAppCtxData(apiConfig))
	api.Use(injectEncryptedMetadata(apiConfig.MetadataEncryptionSecret))
	api.Use(authenticateClient)
	api.Post("/message", apiPostMessageHandlerWithoutStatus)
	api.Post("/message/with-status", apiPostMessageHandlerWithStatus)
	return api
}

func sendJsonResponseString(c *fiber.Ctx, statusCode int, result string) error {
	message := struct {
		Result string `json:"result"`
	}{
		Result: result,
	}
	payload, _ := json.Marshal(message)
	c.Response().Header.SetContentType(fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Status(statusCode).Send(payload)
}

func injectAppCtxData(data *APIConfig) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		c.Locals(apiCtxConfigKey, data)
		return c.Next()
	}
}

func injectEncryptedMetadata(encryptionKey string) func(*fiber.Ctx) error {
	aesKey := sha256.Sum256([]byte(encryptionKey))
	return func(c *fiber.Ctx) error {
		err := storeEncryptedMetadataInCtx(c, aesKey)
		if err != nil {
			return fmt.Errorf("unable to inject encrypted metadata: %w", err)
		}
		return c.Next()
	}
}

func extractAppCtxData(c *fiber.Ctx) *APIConfig {
	return c.Locals(apiCtxConfigKey).(*APIConfig)
}

func apiPostMessageHandlerWithStatus(c *fiber.Ctx) error {
	return apiPostMessageHandler(c, true)
}

func apiPostMessageHandlerWithoutStatus(c *fiber.Ctx) error {
	return apiPostMessageHandler(c, false)
}

func apiPostMessageHandler(c *fiber.Ctx, requireStatus bool) error {
	ctxData := extractAppCtxData(c)
	b := ctxData.Bot

	var message Message
	if err := c.BodyParser(&message); err != nil {
		return sendJsonResponseString(c, fiber.StatusUnprocessableEntity, "unable to parse json")
	}
	mailContact, err := mail.ParseAddress(message.To)
	if err != nil {
		return sendJsonResponseString(c, fiber.StatusUnprocessableEntity, "unable to parse mail address")
	}

	addr := strings.ToLower(mailContact.Address)

	if ctxData.CheckAllowedSend != nil {
		err = ctxData.CheckAllowedSend(addr)
		if err != nil {
			return sendJsonResponseString(c, fiber.StatusUnavailableForLegalReasons, err.Error())
		}
	}

	var chatId uuid.UUID
	// Decide if it is UUID-Group-like or not:
	addrPrefix, ok := strings.CutSuffix(addr, ctxData.GroupChatMailSuffix)
	if !ok {
		users, err := b.FindUsersByMails([]string{addr})
		if err != nil {
			return sendJsonResponseString(c, fiber.StatusInternalServerError, err.Error())
		}
		if len(users) <= 0 {
			return sendJsonResponseString(c, fiber.StatusNotFound, "no users found")
		}
		if len(users) != 1 {
			return sendJsonResponseString(c, fiber.StatusExpectationFailed, "found more than one recepients")
		}
		if users[0].UserKind != "cts_user" {
			return sendJsonResponseString(c, fiber.StatusPreconditionRequired, "user is not cts_user")
		}
		chatId, err = b.CreateChatWithUser(users[0])
		if err != nil {
			return sendJsonResponseString(c, fiber.StatusServiceUnavailable, err.Error())
		}
	} else {
		chatId, err = uuid.Parse(addrPrefix)
		if chatId.String() != addrPrefix {
			sendJsonResponseString(c, fiber.StatusUnprocessableEntity, fmt.Sprintf("chat_id '%s' in address '%s' is not recognized as UUID: %s", addrPrefix, mailContact, err.Error()))
		}
	}

	if ctxData.CheckAllowedSend != nil {
		err = ctxData.CheckAllowedSend(chatId.String())
		if err != nil {
			return sendJsonResponseString(c, fiber.StatusUnavailableForLegalReasons, err.Error())
		}
	}

	// Create Message

	ndOpts := []models.NDRequestOption{}
	if len(message.Buttons) > 0 {
		for _, row := range message.Buttons {
			if len(row) > 0 {
				ndButtonRow := models.NDButtonRow{}
				for _, button := range row {
					opts := []models.NDButtonOption{}
					if button.TextAlign != "" {
						opts = append(opts, models.WithButtonContentAlign(models.NDButtonAlign(button.TextAlign)))
					}
					if button.TextColor != "" {
						opts = append(opts, models.WithButtonFontColor(button.TextColor))
					}
					if button.BackgroundColor != "" {
						opts = append(opts, models.WithButtonBackgroundColor(button.BackgroundColor))
					}
					if button.AlertText != "" {
						opts = append(opts, models.WithButtonAlert(button.AlertText))
					}
					if button.HorizontalSize != 0 {
						opts = append(opts, models.WithButtonHorizontalSize(button.HorizontalSize))
					}
					ndButton := models.NewLinkButton(button.Label, button.Link, opts...)
					ndButtonRow = append(ndButtonRow, ndButton)
				}
				ndOpts = append(ndOpts, models.WithNDBubbleRow(ndButtonRow...))
			}
		}
	}

	ndOpts = append(ndOpts, models.WithNDMetadata(loadEncryptedMetadataFromCtx(c)))

	ndr, err := models.NewNDRequest(chatId, message.Body, ndOpts...)
	if err != nil {
		return err
	}

	if !requireStatus {
		_, err = b.SendMessageAsync(ndr)
		if err != nil {
			return sendJsonResponseString(c, fiber.StatusServiceUnavailable, err.Error())
		}
		return sendJsonResponseString(c, fiber.StatusAccepted, "OK")
	} else {
		_, err = b.SendMessageSync(ndr)
		if err != nil {
			return sendJsonResponseString(c, fiber.StatusServiceUnavailable, err.Error())
		}
		return sendJsonResponseString(c, fiber.StatusCreated, "OK")
	}
}

func authenticateClient(c *fiber.Ctx) error {
	ctxData := extractAppCtxData(c)
	if ctxData.CheckBearerToken == nil {
		return errors.New("token check function not configured")
	}
	tokenString := extractBearerToken(c.Get(fiber.HeaderAuthorization, ""))
	if tokenString == "" {
		return sendJsonResponseString(c, fiber.StatusUnauthorized, "required 'Authorization: Bearer <token>' header")
	}
	if err := ctxData.CheckBearerToken(tokenString); err != nil {
		return sendJsonResponseString(c, fiber.StatusUnauthorized, "provided token is not authorized")
	}
	return c.Next()
}

func extractBearerToken(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)

	if len(authHeader) < 7 || !strings.EqualFold(authHeader[:6], "bearer") {
		return ""
	}
	tokenPart := authHeader[6:]

	token := strings.TrimSpace(tokenPart)

	if token == "" {
		return ""
	}

	return token
}
