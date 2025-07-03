package utils

import "github.com/fatemehkarimi/chronos_bot/entities"

func GetCalendarByType(cType entities.CalendarType) entities.Calendar {
	switch cType {
	case entities.KhorshidiCalendarType:
		return entities.KhorshidiCalendar{}
	case entities.QamariCalendarType:
		return entities.QamariCalendar{}
	default:
		return entities.GeorgianCalendar{}
	}
}
