package utils

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/fatemehkarimi/chronos_bot/entities"
)

func ParseSchedulePattern(pattern string) (*entities.Schedule, error) {
	scheduleKeys := map[string]bool{
		"y":  false,
		"m":  false,
		"d":  false,
		"hh": false,
		"mm": false,
	}
	result := make(map[string]int)
	lines := strings.Split(pattern, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		if _, ok := scheduleKeys[key]; ok {
			scheduleKeys[key] = true
		} else {
			continue
		}

		if valueStr == "#" {
			result[key] = 0
		} else if num, err := strconv.Atoi(valueStr); err == nil {
			result[key] = num
		}
	}

	var schedule entities.Schedule
	for k, _ := range scheduleKeys {
		switch k {
		case "y":
			schedule.Calendar.Year = result["y"]
		case "m":
			schedule.Calendar.Month = result["m"]
		case "d":
			schedule.Calendar.Day = result["d"]
		case "hh":
			schedule.Calendar.Hour = result["hh"]
		case "mm":
			schedule.Calendar.Minute = result["mm"]
		}
	}

	if schedule.Calendar.Day == 0 {
		return nil, fmt.Errorf("مقداری برای روز نیست. لطفا (d) را بفرستید")
	}

	return &schedule, nil
}

func ScheduleTaskOnSameDay(
	dayTime entities.CalendarTime,
	task func() error,
) error {
	now := time.Now()
	scheduleTime := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		dayTime.Hour,
		dayTime.Minute,
		0,
		0,
		now.Location(),
	)
	diff := scheduleTime.Sub(now)

	timer := time.NewTimer(diff)
	tick := <-timer.C

	slog.Info("performing task at time = ", slog.Time("time", tick))
	err := task()
	return err
}

func ScheduleToText(schedule entities.Schedule) string {
	template := `پرچم: %s
گروه کاربران: %s
مقدار: %s
`
	return fmt.Sprintf(
		template,
		schedule.FeatureFlagName,
		schedule.UsersList,
		schedule.Value,
	)
}

func ShouldRunToday(
	calendar entities.Calendar,
	schedule entities.Schedule,
) bool {
	now := calendar.GetToday()
	year := now.Year
	month := now.Month
	day := now.Day
	hour := time.Now().Hour()
	minute := time.Now().Minute()

	return day == schedule.Calendar.Day &&
		(schedule.Calendar.Month == 0 || schedule.Calendar.Month == month) &&
		(schedule.Calendar.Year == 0 || schedule.Calendar.Year == year) &&
		((schedule.Calendar.Hour == hour && schedule.Calendar.Minute >= minute) ||
			(schedule.Calendar.Hour > hour))
}
