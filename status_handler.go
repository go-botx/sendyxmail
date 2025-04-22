package main

import (
	"github.com/go-botx/botx"
	"github.com/go-botx/botx/models"
)

func NewStatusHandler() botx.StatusCallbackHandler {
	return func(b *botx.Bot, req *models.StatusRequest) *models.StatusResponse {
		if req.ChatType != models.ChatTypeChat && !req.IsAdmin {
			return nil
		}
		return models.NewStatusResponse(true, "",
			commandMute,
			commandUnmute)
	}
}
