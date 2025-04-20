package main

import (
	"github.com/go-botx/bot"
	"github.com/go-botx/bot/models"
)

func NewStatusHandler() bot.StatusCallbackHandler {
	return func(b *bot.Bot, req *models.StatusRequest) *models.StatusResponse {
		if !req.IsAdmin {
			return nil
		}
		return models.NewStatusResponse(true, "",
			commandMute,
			commandUnmute)
	}
}
