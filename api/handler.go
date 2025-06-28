package api

import (
	"encoding/json"
	"fmt"
	"github.com/fatemehkarimi/chronos_bot/pkg/utils"
	"log/slog"
	"net/http"
	"strings"

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

func (h *HttpHandler) GetLastProcessedUpdateId() int {
	return h.updateId
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

func (h *HttpHandler) HandleMessageUpdate(updateId int, message *entities.Message) {
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
	fmt.Println("here user state name = ", userState.StateName)
	switch userState.StateName {
	case entities.AddFeatureFlagState:
		h.AddFeatureFlag(updateId, chatId, message)
	case entities.GetScheduleState:
		h.HandleGetSchedule(updateId, int(chatId), *message)
	case entities.GetValueState:
		h.HandleGetValues(updateId, int(chatId), *message)
	case entities.GetUserListState:
		h.HandleUsersList(updateId, int(chatId), *message)
	default:
		slog.Error("unhandled default case", slog.Int("stateName", int(userState.StateName)))
	}
}

func (h *HttpHandler) HandleCallbackQueryUpdate(updateId int, callbackQuery *entities.CallbackQuery) {
	data := callbackQuery.Data
	switch {
	case *data == utils.AddFeatureFlagCallbackData:
		h.HandleAddFeatureFlagCallbackData(updateId, callbackQuery.From.Id)
	case *data == utils.AddScheduleCallbackData:
		h.HandleAddScheduleCallbackData(updateId, callbackQuery.From.Id)
	case *data == utils.KhorshidiCalendarCallbackData:
		h.HandleCalendarTypeCallbackData(updateId, callbackQuery.From.Id, utils.CallbackDataToCalendarType(*data))
	case *data == utils.GeorgianCalendarCallbackData:
		h.HandleCalendarTypeCallbackData(updateId, callbackQuery.From.Id, utils.CallbackDataToCalendarType(*data))
	case *data == utils.QamariCalendarCallbackData:
		h.HandleCalendarTypeCallbackData(updateId, callbackQuery.From.Id, utils.CallbackDataToCalendarType(*data))
	case strings.HasPrefix(*data, "feature_flag"):
		h.HandleChooseFeatureFlag(updateId, callbackQuery.From.Id, *data)
	case *data == utils.UsersListForAllCallbackData:
		value := "*"
		h.HandleUsersList(updateId, callbackQuery.From.Id, entities.Message{Text: &value})
	default:
		slog.Info("unknown callback query data", data)
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
	h.HandleSendUserFeatureFlags(updateId, chatId)
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

func (h *HttpHandler) HandleSendUserFeatureFlags(updateId, chatId int) {
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

	schedule := userState.Schedule
	if schedule != nil {
		schedule.CalendarType = calendarType

		h.userStates[fmt.Sprint(chatId)] = entities.UserState{
			StateName: entities.GetScheduleState,
			Schedule:  schedule,
		}
	}

	userState.StateName = entities.GetScheduleState
	chMessage := make(chan entities.MethodResponse)

	go h.api.SendMessage(
		fmt.Sprint(chatId),
		`برنامه زمانی پرچم را با الگوی زیر بفرستید. برای پارامترهای روز(d)، ساعت(hh) و دقیقه(mm) باید مقداری تعیین شود اما پارامترهای دیگر می‌توانند خالی باشند. اگر به راهنمایی بیشتر نیاز دارید، /help را بفرستید
y:
m:
d: 20
hh: 0
mm: 30
`,
		nil,
		chMessage)

	result := <-chMessage
	if result.Err != nil {
		slog.Error("error send schedule message to user", slog.Int("chatId", chatId))
		h.ResetUserStateAndSendResetMessage(chatId)
		return
	}
}

func (h *HttpHandler) ResetUserStateAndSendResetMessage(chatId int) {
	chFailed := make(chan entities.MethodResponse)
	go h.api.SendMessage(fmt.Sprint(chatId), "خطایی رخ داده است. لطفا دوباره /start را بفرستید", nil, chFailed)
	h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
}

func (h *HttpHandler) SendContactDeveloperErrorMessage(updateId, chatId int) {
	chMessage := make(chan entities.MethodResponse)
	replyMarkup := utils.GetMainReplyMarkup()
	go h.api.SendMessage(fmt.Sprint(chatId), "خطای نامشخص رخ داده است. این موضوع را با توسعه دهنده در میان بگذارید.", replyMarkup, chMessage)

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

func (h *HttpHandler) HandleChooseFeatureFlag(updateId, chatId int, featureFlagCallbackData string) {
	featureFlagName := strings.TrimPrefix(featureFlagCallbackData, "feature_flag ")
	userState := h.userStates[fmt.Sprint(chatId)]

	if userState.StateName != entities.ChooseFeatureFlagState {
		h.ResetUserStateAndSendResetMessage(chatId)
		return
	}

	newFeatureFlag := entities.FeatureFlag{Name: featureFlagName, OwnerId: chatId}
	h.HandleSendCalendarType(updateId, chatId, &newFeatureFlag)
}

func (h *HttpHandler) HandleSendCalendarType(updateId, chatId int, featureFlag *entities.FeatureFlag) {
	ch := make(chan entities.MethodResponse)
	replyMarkup := utils.GetScheduleReplyMarkup()
	go h.api.SendMessage(fmt.Sprint(chatId), "تقویم برنامه زمانی را انتخاب کنید", replyMarkup, ch)
	result := <-ch
	if result.Err != nil {
		slog.Error("error send choose calendar message. err = ", slog.Int("chatId", chatId), slog.Any("err", result.Err))
		h.ResetUserStateAndSendResetMessage(chatId)
		return
	} else {
		h.userStates[fmt.Sprint(chatId)] = entities.UserState{
			StateName: entities.ChooseCalendarTypeState,
			Schedule: &entities.Schedule{
				FeatureFlagName: featureFlag.Name,
			},
		}
	}
}

func (h *HttpHandler) HandleGetSchedule(updateId, chatId int, message entities.Message) {
	text := message.Text
	schedule, err := utils.ParseSchedulePattern(*text)

	if err != nil {
		go h.api.SendMessage(fmt.Sprint(chatId), err.Error(), nil, nil)
		return
	}

	userSchedule := h.userStates[fmt.Sprint(chatId)].Schedule
	if userSchedule != nil {
		userSchedule.Year = schedule.Year
		userSchedule.Month = schedule.Month
		userSchedule.Day = schedule.Day
		userSchedule.Hour = schedule.Hour
		userSchedule.Minute = schedule.Minute
		h.userStates[fmt.Sprint(chatId)] = entities.UserState{
			StateName: entities.GetValueState,
			Schedule:  userSchedule,
		}
	} else {
		h.SendContactDeveloperErrorMessage(updateId, chatId)
		slog.Error("user is in userSchedule state but not feature flag value is provided", slog.Int("updateId", updateId), slog.Int("chatId", chatId))
		return
	}
	go h.api.SendMessage(fmt.Sprint(chatId), "مقدار(value) پرچم را وارد کنید. با اجرای برنامه زمانی، این مقدار برای کاربران تنظیم می شود.", nil, nil)
}

func (h *HttpHandler) HandleGetValues(updateId, chatId int, message entities.Message) {
	value := message.Text
	userState := h.userStates[fmt.Sprint(chatId)]
	schedule := userState.Schedule
	if schedule != nil {
		schedule.Value = *value
		h.userStates[fmt.Sprint(chatId)] = entities.UserState{
			StateName: entities.GetUserListState,
			Schedule:  schedule,
		}
	}

	ch := make(chan entities.MethodResponse)
	replyMarkup := utils.GetUsersListCReplyMarkup()
	go h.api.SendMessage(fmt.Sprint(chatId), "شناسه‌ی کاربری(id) کاربرانی که می‌خواهید برای آنها این مقدار تنظیم شود را بنویسید. برای تنظیم مقدار برای همه‌ی کاربران، روی دکمه 'همه کاربران' کلیک کنید. ", replyMarkup, ch)
}

func (h *HttpHandler) HandleUsersList(updateId, chatId int, message entities.Message) {
	value := message.Text
	userState := h.userStates[fmt.Sprint(chatId)]
	schedule := userState.Schedule
	fmt.Println("here value = ", value)
	if schedule != nil {
		schedule.UsersList = *value
		scheduleId, err := h.db.AddSchedule(*schedule)

		if err != nil {
			h.ResetUserStateAndSendResetMessage(chatId)
			slog.Error("error save schedule. err = ", slog.Any("error", err))
			return
		}

		h.userStates[fmt.Sprint(chatId)] = entities.UserState{StateName: entities.StartState}
		replyMarkup := utils.GetMainReplyMarkup()
		go h.api.SendMessage(fmt.Sprint(chatId), fmt.Sprintf("برنامه زمانی با شناسه %d با موفقیت ذخیره شد", scheduleId), replyMarkup, nil)
	}
}
