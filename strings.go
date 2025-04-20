package main

import "strings"

var (
	localizedMessages = map[string]string{
		"ru-not_admin":           "Только админы чата могут управлять мной.",
		"ru-muted":               "Я не буду доставлять сообщения в этот чат.",
		"ru-unmuted":             "Я буду доставлять сообщения в этот чат.",
		"ru-not_changed_muted":   "Я уже отключен.",
		"ru-not_changed_unmuted": "Я доставлю сообщения в этот чат как только их кто-то отправит.",
		"ru-show_chat_addr":      "Адрес данного чата для отправки сообщений через бота: `%s`",
		"ru-error":               "Что-то пошло не так...",

		"en-not_admin":           "Only chat admins can control me.",
		"en-muted":               "I stopped delivering messages to this chat.",
		"en-unmuted":             "I started delivering messages to this chat.",
		"en-not_changed_muted":   "I already stopped delivering messages to this chat.",
		"en-not_changed_unmuted": "I will deliver messages to this chat as soon as someone sends them.",
		"en-show_chat_addr":      "The address of this chat for sending messages via the bot: `%s`",
		"en-error":               "Something is wrong...",
	}
)

const defaultLocale = string("ru")

func getLocalizedMessage(locale string, messageId string) string {
	locale = strings.ToLower(locale)
	messageId = strings.ToLower(messageId)
	if msg, ok := localizedMessages[locale+"-"+messageId]; ok {
		return msg
	}
	if msg, ok := localizedMessages[defaultLocale+"-"+messageId]; ok {
		return msg
	}
	return "{{" + locale + "-" + messageId + "}}"
}
