package utils

import "github.com/fatemehkarimi/chronos_bot/entities"

const (
	AddFeatureFlagCallbackData = "add feature_flag"
	AddScheduleCallbackData    = "add schedule"
	ReturnCallbackData         = "return"
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
