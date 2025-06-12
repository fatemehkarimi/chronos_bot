package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/repository"
)

type Handler interface {
	GetUpdates(w http.ResponseWriter, r *http.Request)
}

type HttpHandler struct {
	db       repository.Repository
	api      Api
	updateId int
}

func NewHttpHandler(db repository.Repository, token string) Handler {
	api := BaleApi{token: token}
	return &HttpHandler{db: db, api: api}
}

func (h *HttpHandler) GetUpdates(w http.ResponseWriter, r *http.Request) {
	var update entities.Update
	err := json.NewDecoder(r.Body).Decode(&update)

	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.WriteHeader(http.StatusOK)

	if update.Message != nil {
		h.HandleMessageUpdate(update.UpdateId, update.Message)
	}
}

func (h *HttpHandler) HandleMessageUpdate(updatedId int, message *entities.Message) {
	from := message.From
	chat := message.Chat
	if from.IsBot || chat.Type != "private" {
		return
	}

	text := message.Text
	if text == nil {
		return
	}

	if *text == "/start" {
		chatId := chat.Id
		ch := make(chan entities.MethodResponse)

		scheduleCallbackData := "add schedule"
		featureFlagCallbackData := "add feature_flag"
		replyMarkup := entities.InlineKeyboardMarkup{
			InlineKeyboard: [][]entities.InlineKeyboardButton{
				{
					entities.InlineKeyboardButton{Text: "افزودن پرچم", CallbackData: &featureFlagCallbackData},
				},
				{
					entities.InlineKeyboardButton{Text: "افزودن برنامه زمانی", CallbackData: &scheduleCallbackData},
				},
			},
		}

		go h.api.SendMessage(fmt.Sprint(chatId), "سلام!\nچه کاری را می خواهید به من بسپارید؟", &replyMarkup, ch)

		result := <-ch

		if result.Err != nil {
			fmt.Println("failed to send response", result.Err)
			return
		} else {
			h.updateId = max(h.updateId, updatedId)
		}

		fmt.Println(result.Response)
	}
}
