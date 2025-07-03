package utils

import (
	"fmt"
	"github.com/fatemehkarimi/chronos_bot/entities"
)

const (
	AddFeatureFlagCallbackData = "add feature_flag"
	AddScheduleCallbackData    = "add scheduler"

	KhorshidiCalendarCallbackData = "khorshidi calendar"
	GeorgianCalendarCallbackData  = "georgian calendar"
	QamariCalendarCallbackData    = "qamari calendar"

	UsersListForAllCallbackData = "usersList for all"
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

func GetReplyMarkupFromFeatureFlags(featureFlags []entities.FeatureFlag) entities.ReplyMarkup {
	inlineKeyboard := make([][]entities.InlineKeyboardButton, len(featureFlags))
	for idx, featureFlag := range featureFlags {
		callbackData := fmt.Sprintf("feature_flag %s", featureFlag.Name)
		inlineKeyboard[idx] = []entities.InlineKeyboardButton{
			entities.InlineKeyboardButton{Text: featureFlag.Name, CallbackData: &callbackData},
		}
	}
	return entities.InlineKeyboardMarkup{
		InlineKeyboard: inlineKeyboard,
	}
}

func GetUsersListCReplyMarkup() entities.ReplyMarkup {
	usersListCallbackData := UsersListForAllCallbackData
	replyMarkup := entities.InlineKeyboardMarkup{
		InlineKeyboard: [][]entities.InlineKeyboardButton{
			{
				entities.InlineKeyboardButton{Text: "همه‌ی کاربران", CallbackData: &usersListCallbackData},
			},
		},
	}
	return replyMarkup
}
