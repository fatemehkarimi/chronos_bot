package utils

import "github.com/fatemehkarimi/chronos_bot/entities"

const (
	AddFeatureFlagCallbackData = "add feature_flag"
	AddScheduleCallbackData    = "add schedule"

	KhorshidiCalendarCallbackData = "khorshidi calendar"
	GeorgianCalendarCallbackData  = "georgian calendar"
	QamariCalendarCallbackData    = "qamari calendar"
)

func GetMainReplyMarkup() entities.ReplyMarkup {
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
	return replyMarkup
}

func GetScheduleReplyMarkup() entities.ReplyMarkup {
	khorshidiCalendarCallbackData := KhorshidiCalendarCallbackData
	georgianCalendarCallbackData := GeorgianCalendarCallbackData
	qamariCalendarCallbackData := QamariCalendarCallbackData

	replyMarkup := entities.InlineKeyboardMarkup{
		InlineKeyboard: [][]entities.InlineKeyboardButton{
			{
				entities.InlineKeyboardButton{Text: "خورشیدی", CallbackData: &khorshidiCalendarCallbackData},
			},
			{
				entities.InlineKeyboardButton{Text: "میلادی", CallbackData: &georgianCalendarCallbackData},
			},
			{
				entities.InlineKeyboardButton{Text: "قمری", CallbackData: &qamariCalendarCallbackData},
			},
		},
	}

	return replyMarkup
}
