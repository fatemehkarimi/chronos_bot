package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatemehkarimi/chronos_bot/entities"
)

type Api interface {
	SendMessage(
		chatId string,
		text string,
		replyMarkUp entities.ReplyMarkup,
	) entities.MethodResponse
}

type BaleApi struct {
	token string
}

func NewBaleApi(token string) BaleApi {
	return BaleApi{token}
}

func (api BaleApi) SendMessage(
	chatId string,
	text string,
	replyMarkUp entities.ReplyMarkup,
) entities.MethodResponse {
	requestStruct := entities.RequestSendMessage{
		ChatId:      chatId,
		Text:        text,
		ReplyMarkup: &replyMarkUp,
	}

	requestBytes, err := json.MarshalIndent(requestStruct, "", "  ")
	if err != nil {
		return entities.MethodResponse{
			Err: err,
		}
	}
	endpoint := fmt.Sprintf("https://tapi.bale.ai/bot%s/sendMessage", api.token)
	req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer(requestBytes))

	if err != nil {
		return entities.MethodResponse{
			Err: err,
		}
	}

	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		return entities.MethodResponse{
			Err: err,
		}
	}

	return entities.MethodResponse{Response: res, Err: nil}
}
