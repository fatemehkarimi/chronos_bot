package api

import (
	"encoding/json"
	"fmt"
	"github.com/fatemehkarimi/chronos_bot/pkg/utils"
	"log/slog"
	"net/http"

	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/repository"
	"github.com/lib/pq"
)

type Handler interface {
	GetUpdates(w http.ResponseWriter, r *http.Request)
	GetLastProcessedUpdateId() int
}

type HttpHandler struct {
	db         repository.Repository
	api        Api
	updateId   int
	userStates map[string]entities.UserState
}

func NewHttpHandler(db repository.Repository, token string) Handler {
	api := BaleApi{token: token}
	return &HttpHandler{db: db, api: api, userStates: map[string]entities.UserState{}, updateId: 57}
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

	h.updateId = max(h.updateId, update.UpdateId)
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

		h.userStates[fmt.Sprintf("%d", chatId)] = entities.UserState{StateName: entities.StartState}
		ch := make(chan entities.MethodResponse)

		replyMarkup := utils.GetMainReplyMarkup()
		go h.api.SendMessage(fmt.Sprint(chatId), "سلام!\nچه کاری را می خواهید به من بسپارید؟", replyMarkup, ch)

		result := <-ch

		if result.Err != nil {
			fmt.Println("failed to send response", result.Err)

			chFailed := make(chan entities.MethodResponse)
			slog.Error("error handling /start command. err = ", slog.Int64("chatId", chatId), slog.Any("err", result.Err))
			go h.api.SendMessage(fmt.Sprint(chatId), "خطایی رخ داده است. لطفا دوباره /start را بفرستید", nil, chFailed)
			return
		} else {
		}

		fmt.Println(result.Response)
		return
	}

	userState := h.userStates[fmt.Sprint(chatId)]
	switch userState.StateName {
	case entities.AddFeatureFlagState:
		h.AddFeatureFlag(updatedId, chatId, message)
		return
	default:
		slog.Error("unhandled default case", slog.Int("stateName", int(userState.StateName)))
	}
}

func (h *HttpHandler) HandleCallbackQueryUpdate(updateId int, callbackQuery *entities.CallbackQuery) {
	data := callbackQuery.Data
	switch *data {
	case utils.AddFeatureFlagCallbackData:
		h.HandleAddFeatureFlagCallbackData(updateId, callbackQuery.From.Id)
	case utils.AddScheduleCallbackData:
		h.HandleAddScheduleCallbackData(updateId, callbackQuery.From.Id)
	case utils.KhorshidiCalendarCallbackData:
	case utils.GeorgianCalendarCallbackData:
	case utils.QamariCalendarCallbackData:
		h.HandleCalendarTypeCallbackData(updateId, callbackQuery.From.Id, utils.CallbackDataToCalendarType(*data))
	default:
		slog.Info("unknown callback query data", data)
		return
	}
}

func (h *HttpHandler) HandleAddFeatureFlagCallbackData(updateId, chatId int) {
	ch := make(chan entities.MethodResponse)
	go h.api.SendMessage(fmt.Sprint(chatId), "نام پرچم(feature flag) را بنویسید.", nil, ch)

	result := <-ch
	if result.Err != nil {
		slog.Error("error handling /start command. err = ", slog.Int("chatId", chatId), slog.Any("err", result.Err))
		h.ResetUserStateAndSendResetMessage(chatId)
		return
	} else {
		h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.AddFeatureFlagState}
	}
}

func (h *HttpHandler) HandleAddScheduleCallbackData(updateId int, chatId int) {
	//ch := make(chan entities.MethodResponse)
	//replyMarkup := utils.GetScheduleReplyMarkup()
	//go h.api.SendMessage(fmt.Sprint(chatId), "تقویم برنامه زمانی را انتخاب کنید", replyMarkup, ch)
	//result := <-ch
	//if result.Err != nil {
	//	slog.Error("error send choose calendar message. err = ", slog.Int("chatId", chatId), slog.Any("err", result.Err))
	//	h.ResetUserStateAndSendResetMessage(chatId)
	//	return
	//} else {
	//	h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.ChooseFeatureFlagState}
	//}

	h.HandleChooseFeatureFlag(updateId, chatId)
}

func (h *HttpHandler) AddFeatureFlag(updateId int, chatId int64, message *entities.Message) {
	value := *message.Text

	if value != "" {
		// because chatId is private, casting is fine
		err := h.db.AddFeatureFlag(int(chatId), value)
		chMessage := make(chan entities.MethodResponse)
		if err != nil {
			replyMarkup := utils.GetMainReplyMarkup()
			if pgErr, ok := err.(*pq.Error); ok {
				if pgErr.Code == "23505" && pgErr.Constraint == "feature_flag_pkey" {
					slog.Error("Duplicate key error on feature_flag_pkey",
						slog.Int("updateId", updateId),
						slog.Int64("chatId", chatId),
						slog.String("value", value),
					)

					featureFlag, err := h.db.GetFeatureFlagByName(value)
					if err != nil {
						slog.Error("error getting feature flag", slog.Any("error", err))
					}

					text := "این پرچم به نام کاربر دیگری ثبت شده است."
					if featureFlag.OwnerId == int(chatId) {
						text = "این پرچم قبلا به نام شما ثبت شده است."
					}
					go h.api.SendMessage(fmt.Sprint(chatId), text, replyMarkup, chMessage)

					result := <-chMessage
					if result.Err != nil {
						slog.Error("faild to notify user for duplicate response",
							slog.Int("updateId", updateId),
							slog.Int64("chatId", chatId),
							slog.String("value", value),
						)
						h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
					}
				}
			} else {
				h.SendContactDeveloperErrorMessage(updateId, int(chatId))
			}
		} else {
			go h.api.SendMessage(
				fmt.Sprint(chatId),
				"پرچم شما ثبت شد. اکنون می‌توانید برنامه زمانی برای آن تعریف کنید.",
				utils.GetMainReplyMarkup(),
				chMessage,
			)

			result := <-chMessage
			if result.Err != nil {
				slog.Error("unknown error occurred adding new feature flag",
					slog.Int("updateId", updateId),
					slog.Int64("chatId", chatId),
					slog.String("value", value),
					slog.Any("error", result.Err),
				)
				h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
			}
		}
	}
}

func (h *HttpHandler) HandleChooseFeatureFlag(updateId, chatId int) {
	featureFlags, err := h.db.GetFeatureFlagsByOwnerId(chatId)
	if err != nil {
		slog.Error("error getting feature flags", slog.Any("error", err))
		h.SendContactDeveloperErrorMessage(updateId, chatId)
		return
	}

	if len(featureFlags) == 0 {
		replyMarkup := utils.GetMainReplyMarkup()
		go h.api.SendMessage(fmt.Sprint(chatId), "پرچمی به نام شما ثبت نشده است. پرچم را ثبت کنید تا بتوانید برنامه زمانی برای آن تنظیم کنید.", replyMarkup, nil)
		h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
		return
	}

	chMessage := make(chan entities.MethodResponse)
	replyMarkup := utils.GetReplyMarkupFromFeatureFlags(featureFlags)
	go h.api.SendMessage(fmt.Sprint(chatId), "پرچم را انتخاب کنید.", replyMarkup, chMessage)

	result := <-chMessage
	if result.Err != nil {
		slog.Error("failed to send feature flags to user", slog.Int("updateId", updateId), slog.Int("chatId", chatId), slog.Any("error", result.Err))
		return
	}
	h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.ChooseFeatureFlagState}
}

func (h *HttpHandler) HandleCalendarTypeCallbackData(updateId, chatId int, calendarType entities.Calendar) {
	userState := h.userStates[fmt.Sprint(chatId)]
	if userState.StateName != entities.ChooseCalendarTypeState {
		slog.Error("user cannot set calendar type in this state", slog.Int("updateId", updateId), slog.Int("chatId", chatId), slog.Any("state", userState.StateName))
		h.ResetUserStateAndSendResetMessage(chatId)
		return
	}

	userState.SelectedCalendarType = &calendarType

}

func (h *HttpHandler) ResetUserStateAndSendResetMessage(chatId int) {
	chFailed := make(chan entities.MethodResponse)
	go h.api.SendMessage(fmt.Sprint(chatId), "خطایی رخ داده است. لطفا دوباره /start را بفرستید", nil, chFailed)
	h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
}

func (h *HttpHandler) SendContactDeveloperErrorMessage(updateId, chatId int) {
	chMessage := make(chan entities.MethodResponse)
	replyMarkup := utils.GetMainReplyMarkup()
	go h.api.SendMessage(fmt.Sprint(chatId), "خطای نامشخص در افزودن پرچم رخ داده است. این موضوع را با توسعه دهنده در میان بگذارید.", replyMarkup, chMessage)

	result := <-chMessage
	if result.Err != nil {
		slog.Error("unknown error occurred adding new feature flag",
			slog.Int("updateId", updateId),
			slog.Int("chatId", chatId),
			slog.Any("error", result.Err),
		)
		h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
	}
}

func (h *HttpHandler) GetLastProcessedUpdateId() int {
	return h.updateId
}
