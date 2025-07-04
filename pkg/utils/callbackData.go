package utils

import (
	"github.com/fatemehkarimi/chronos_bot/entities"
	"strings"
)

func CallbackDataToCalendarType(data string) entities.CalendarType {
	switch data {
	case GeorgianCalendarCallbackData:
		return entities.GeorgianCalendarType
	case QamariCalendarCallbackData:
		return entities.QamariCalendarType
	default:
		return entities.KhorshidiCalendarType
	}
}

func GetFeatureFlagNameFromCallbackData(data string) string {
	featureFlagName := strings.TrimPrefix(
		data,
		"feature_flag ",
	)
	return featureFlagName
}
