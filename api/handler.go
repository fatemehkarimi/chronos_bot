package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/repository"
	"github.com/lib/pq"
)

type UserState int

const (
	_ UserState = iota
	StartState
	AddFeatureFlagState
)

const (
	AddFeatureFlagCallbackData = "add feature_flag"
	AddScheduleCallbackData    = "add schedule"
	ReturnCallbackData         = "return"
)

type Handler interface {
	GetUpdates(w http.ResponseWriter, r *http.Request)
	GetLastProcesedUpdateId() int
}

type HttpHandler struct {
	db         repository.Repository
	api        Api
	updateId   int
	userStates map[string]UserState
}

func NewHttpHandler(db repository.Repository, token string) Handler {
	api := BaleApi{token: token}
	return &HttpHandler{db: db, api: api, userStates: map[string]UserState{}, updateId: 28}
}

func (h *HttpHandler) GetUpdates(w http.ResponseWriter, r *http.Request) {
	var update entities.Update
	err := json.NewDecoder(r.Body).Decode(&update)

	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		slog.Error("error parsing update", slog.Any("error", err))
		return
	}

	w.WriteHeader(http.StatusOK)
	if h.updateId >= update.UpdateId {
		return
	}

	if update.Message != nil {
		h.HandleMessageUpdate(update.UpdateId, update.Message)
	}

	if update.CallbackQuery != nil {
		h.HandleCallbackQueryUpdate(update.UpdateId, update.CallbackQuery)
	}
}

func (h *HttpHandler) HandleMessageUpdate(updatedId int, message *entities.Message) {
	chat := message.Chat
	if chat.Type != "private" {
		return
	}

	chatId := chat.Id
	text := message.Text
	if text != nil && *text == "/start" {
		from := message.From
		if from.IsBot {
			return
		}

		h.userStates[fmt.Sprintf("%d", chatId)] = StartState
		ch := make(chan entities.MethodResponse)

		scheduleCallbackData := AddScheduleCallbackData
		featureFlagCallbackData := AddFeatureFlagCallbackData
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

		fmt.Println(result.Response)
	}

	userState := h.userStates[fmt.Sprint(chatId)]
	switch userState {
	case AddFeatureFlagState:
		h.AddFeatureFlag(updatedId, chatId, message)
		return
	}
}

func (h *HttpHandler) HandleCallbackQueryUpdate(updateId int, callbackQuery *entities.CallbackQuery) {
	data := callbackQuery.Data
	switch *data {
	case AddFeatureFlagCallbackData:
		h.HandleAddFeatureFlagCallbackData(updateId, callbackQuery.From.Id)
	default:
		slog.Info("unknown callback query data", data)
		h.updateId = max(h.updateId, updateId)
		return
	}

}

func (h *HttpHandler) HandleAddFeatureFlagCallbackData(updateId, chatId int) {
	ch := make(chan entities.MethodResponse)
	go h.api.SendMessage(fmt.Sprint(chatId), "نام پرچم را بنویسید.(feature flag)", nil, ch)

	result := <-ch
	if result.Err != nil {
		slog.Error("error handling /start command. err = ", slog.Int("chatId", chatId), slog.Any("err", result.Err))
		h.ResetUserStateAndSendResetMessage(chatId)
		return
	} else {
		h.updateId = max(h.updateId, updateId)
		h.userStates[fmt.Sprint(chatId)] = AddFeatureFlagState
	}
}

func (h *HttpHandler) AddFeatureFlag(updateId int, chatId int64, message *entities.Message) {
	value := *message.Text

	if value != "" {
		// because chatId is private, casting is fine
		err := h.db.AddFeatureFlag(int(chatId), value)
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" && pgErr.Constraint == "feature_flag_pkey" {
				slog.Error("Duplicate key error on feature_flag_pkey",
					slog.Int("updateId", updateId),
					slog.Int64("chatId", chatId),
					slog.String("value", value),
				)

				scheduleCallbackData := AddScheduleCallbackData
				featureFlagCallbackData := AddFeatureFlagCallbackData
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

				chMessage := make(chan entities.MethodResponse)
				go h.api.SendMessage(fmt.Sprint(chatId), "این پرچم قبلا به نام شما ثبت شده است.", replyMarkup, chMessage)

				result := <-chMessage
				if result.Err != nil {
					slog.Error("faild to notify user for duplicate response",
						slog.Int("updateId", updateId),
						slog.Int64("chatId", chatId),
						slog.String("value", value),
					)
					h.userStates[fmt.Sprint(chatId)] = StartState
				}
				return
			}
		}
	}
}

func (h *HttpHandler) ResetUserStateAndSendResetMessage(chatId int) {
	chFailed := make(chan entities.MethodResponse)
	go h.api.SendMessage(fmt.Sprint(chatId), "خطایی رخ داده است. لطفا دوباره /start را بفرستید", nil, chFailed)
	h.userStates[fmt.Sprint(chatId)] = StartState
}

func (h *HttpHandler) GetLastProcesedUpdateId() int {
	return h.updateId
}
