package entities

import "net/http"

type MethodResponse struct {
	Response *http.Response
	Err      error
}

type RequestSendMessage struct {
	ChatId           string       `json:"chat_id"`
	Text             string       `json:"text"`
	ReplyToMessageId *int         `json:"reply_to_message_id,omitempty"`
	ReplyMarkup      *ReplyMarkup `json:"reply_markup,omitempty"`
}

type ResponseSendMessage = Message
