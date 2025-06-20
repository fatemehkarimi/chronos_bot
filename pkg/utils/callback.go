package utils

import (
	"github.com/fatemehkarimi/chronos_bot/entities"
)

func CallbackDataToCalendarType(data string) entities.Calendar {
	switch data {
	case GeorgianCalendarCallbackData:
		return entities.GeorgianCalendar
	case QamariCalendarCallbackData:
		return entities.QamariCalendar
	default:
		return entities.KhorshidiCalendar
	}
}
