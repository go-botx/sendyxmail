package main

import "github.com/go-botx/botx/models"

var (
	// Command body MUST be lowercase
	commandMute = models.StatusResponseCommand{
		Body:        "/mute",
		Description: "Mute notifications in this chat 🔕",
	}
	commandUnmute = models.StatusResponseCommand{
		Body:        "/unmute",
		Description: "Unmute notifications in this chat 🔔",
	}
	commandChatAddr = models.StatusResponseCommand{
		Body:        "/_address",
		Description: "Get address for current chat",
	}
)
