package entities

import (
	ptime "github.com/yaa110/go-persian-calendar"
	"time"
)

type Calendar interface {
	Type() CalendarType
	GetToday() CalendarTime
}

type GeorgianCalendar struct{}

func (g GeorgianCalendar) Type() CalendarType {
	return GeorgianCalendarType
}

func (g GeorgianCalendar) GetToday() CalendarTime {
	now := time.Now()
	return CalendarTime{
		Year:   now.Year(),
		Month:  int(now.Month()),
		Day:    now.Day(),
		Hour:   now.Hour(),
		Minute: now.Minute(),
	}
}

type KhorshidiCalendar struct{}

func (k KhorshidiCalendar) Type() CalendarType {
	return KhorshidiCalendarType
}

func (k KhorshidiCalendar) GetToday() CalendarTime {
	now := ptime.Now()
	return CalendarTime{
		Year:   now.Year(),
		Month:  int(now.Month()),
		Day:    now.Day(),
		Hour:   now.Hour(),
		Minute: now.Minute(),
	}
}

type QamariCalendar struct{}

func (q QamariCalendar) Type() CalendarType {
	return QamariCalendarType
}

func (q QamariCalendar) GetToday() CalendarTime {
	panic("not implemented yet")
}
