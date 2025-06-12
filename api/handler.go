package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/repository"
)

type UserState int

const (
	_ UserState = iota
	Start
	AddFeatureFlag
	FailedAddFeatureFlag
	SuccessAddFeatureFlag
)

type Handler interface {
	GetUpdates(w http.ResponseWriter, r *http.Request)
}

type HttpHandler struct {
	db         repository.Repository
	api        Api
	updateId   int
	userStates map[string]UserState
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
	if h.updateId >= update.UpdateId {
		return
	}

	if update.Message != nil {
		h.HandleMessageUpdate(update.UpdateId, update.Message)
	}
}

func (h *HttpHandler) HandleMessageUpdate(updatedId int, message *entities.Message) {
	chat := message.Chat
	if chat.Type != "private" {
		return
	}

	text := message.Text
	if text != nil && *text == "/start" {
		from := message.From
		if from.IsBot {
			return
		}

		chatId := chat.Id
		h.userStates[fmt.Sprintf("%d", chatId)] = Start
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

			chFailed := make(chan entities.MethodResponse)
			slog.Error("error handling /start command. err = ", slog.Int64("chatId", chatId), slog.Any("err", result.Err))
			go h.api.SendMessage(fmt.Sprint(chatId), "خطایی رخ داده است. لطفا دوباره /start را بفرستید", nil, chFailed)
			return
		} else {
			h.updateId = max(h.updateId, updatedId)
		}

		h.userStates[fmt.Sprintf("%d", chatId)] = AddFeatureFlag
		fmt.Println(result.Response)
	}
}
