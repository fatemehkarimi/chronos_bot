package utils

import (
	"fmt"
	"github.com/fatemehkarimi/chronos_bot/entities"
	"strconv"
	"strings"
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
