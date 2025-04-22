package main

import (
	"fmt"
	"log"
	"slices"
	"strings"

	botx "github.com/go-botx/botx"
	"github.com/go-botx/botx/models"
)

func NewCommandHandler() botx.CommandCallbackHandler {
	return func(b *botx.Bot, req *models.CommandRequest) {
		chatId := req.From.GroupChatId
		isAdmin := req.From.IsAdmin || req.From.ChatType == models.ChatTypeChat

		if req.Command.CommandType == models.CommandTypeUser {
			bodyParts := strings.SplitN(strings.TrimSpace(strings.ToLower(req.Command.Body)), " ", 2)
			if len(bodyParts) < 1 {
				return
			}
			command := bodyParts[0]
			if slices.Contains([]string{commandMute.Body, commandUnmute.Body}, command) {
				responseString := ""
				if !isAdmin {
					responseString = "not_admin"
				} else {
					changed, err := mm.SetMute(chatId.String(), command == commandMute.Body)
					if err != nil {
						log.Printf("failed to change mute for %s to %t: %s", chatId, command == commandMute.Body, err.Error())
						responseString = "error"
					} else if !changed {
						if command == commandMute.Body {
							responseString = "not_changed_muted"
						} else {
							responseString = "not_changed_unmuted"
						}

					} else {
						if command == commandMute.Body {
							responseString = "muted"
						} else {
							responseString = "unmuted"
						}
					}
				}

				if responseString != "" {
					message, err := models.NewNDRequest(chatId, getLocalizedMessage(req.From.Locale, responseString))
					if err != nil {
						return
					}
					b.SendMessageAsync(message)
				}
			}
			if command == commandChatAddr.Body && isAdmin {
				message, err := models.NewNDRequest(chatId, fmt.Sprintf(getLocalizedMessage(req.From.Locale, "show_chat_addr"), fmt.Sprintf("%s%s", chatId.String(), groupChatMailSuffix)))
				if err != nil {
					return
				}
				b.SendMessageAsync(message)
			}
		}

	}
}
