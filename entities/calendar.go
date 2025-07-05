package entities

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
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
	res, err := http.Get("https://api.aladhan.com/v1/gToH")
	if err != nil {
		return CalendarTime{}
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return CalendarTime{}
	}

	var cTime CalendarTime
	var resp AladhanDateResponse
	err = json.Unmarshal(body, &resp)

	if err != nil {
		return cTime
	}
	cTime.Year, _ = strconv.Atoi(resp.Data.Hijri.Year)
	cTime.Month = resp.Data.Hijri.Month.Number
	cTime.Day, _ = strconv.Atoi(resp.Data.Hijri.Day)
	return cTime
}
