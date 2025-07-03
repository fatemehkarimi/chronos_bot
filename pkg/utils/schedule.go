package utils

import (
	"fmt"
	"github.com/fatemehkarimi/chronos_bot/entities"
	"log/slog"
	"strconv"
	"strings"
	"time"
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
			schedule.Year = result["y"]
		case "m":
			schedule.Month = result["m"]
		case "d":
			schedule.Day = result["d"]
		case "hh":
			schedule.Hour = result["hh"]
		case "mm":
			schedule.Minute = result["mm"]
		}
	}

	if schedule.Day == 0 {
		return nil, fmt.Errorf("مقداری برای روز نیست. لطفا (d) را بفرستید")
	}

	return &schedule, nil
}

func ScheduleTaskOnSameDay(dayTime entities.DayTime, task func() error) error {
	now := time.Now()
	scheduleTime := time.Date(now.Year(), now.Month(), now.Day(), dayTime.Hour, dayTime.Minute, 0, 0, now.Location())
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
	return fmt.Sprintf(template, schedule.FeatureFlagName, schedule.UsersList, schedule.Value)
}

func ShouldRunToday(schedule entities.Schedule) bool {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()

	return day == schedule.Day &&
		(schedule.Month == 0 || schedule.Month == month) &&
		(schedule.Year == 0 || schedule.Year == year) &&
		((schedule.Hour == hour && schedule.Minute >= minute) ||
			(schedule.Hour > hour))
}
