package scheduler

import (
	"fmt"
	"log/slog"

	"github.com/fatemehkarimi/chronos_bot/api"
	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/pkg/utils"
	"github.com/fatemehkarimi/chronos_bot/repository"
)

type Scheduler interface {
	LaunchSchedulesInRange(
		calendar entities.Calendar,
		startDayTime entities.CalendarTime,
		endDayTime entities.CalendarTime,
	)
	OnNewSchedule(schedule entities.Schedule)
}

type DBScheduler struct {
	repo       repository.Repository
	api        api.Api
	logChannel string
}

func NewScheduler(
	DB repository.Repository,
	api api.Api,
	logChannel string,
) Scheduler {
	return DBScheduler{DB, api, logChannel}
}

func (s DBScheduler) LaunchSchedulesInRange(
	calendar entities.Calendar,
	startDayTime entities.CalendarTime,
	endDayTime entities.CalendarTime,
) {
	now := calendar.GetToday()
	year := now.Year
	month := now.Month
	day := now.Day

	schedules, err := s.repo.GetScheduleByTime(
		calendar.Type(),
		year,
		month,
		day,
		startDayTime,
		endDayTime,
	)

	if err != nil {
		slog.Error("error getting schedules", slog.Any("error", err))
		return
	}

	slog.Info("found schedules", slog.Any("schedules", schedules))

	for _, schedule := range schedules {
		go s.ScheduleAndNotify(schedule)
	}
}

func (s DBScheduler) OnNewSchedule(schedule entities.Schedule) {
	calendar := utils.GetCalendarByType(schedule.Calendar.Type)
	fmt.Println("here on new schedule = ", utils.ShouldRunToday(calendar, schedule))
	if utils.ShouldRunToday(calendar, schedule) {
		go s.ScheduleAndNotify(schedule)
	}
}

func (s DBScheduler) ScheduleAndNotify(schedule entities.Schedule) {
	taskDayTime := entities.CalendarTime{
		Hour:   schedule.Calendar.Hour,
		Minute: schedule.Calendar.Minute,
	}
	task := func() error {
		SetConfig(schedule)
		result := s.api.SendMessage(
			s.logChannel,
			utils.ScheduleToText(schedule),
			nil,
		)

		if result.Err != nil {
			slog.Error(
				"error sending schedule to log channel",
				slog.Any("error", result.Err),
				slog.Any("schedule", schedule),
			)
		}

		return nil
	}
	err := utils.ScheduleTaskOnSameDay(taskDayTime, task)
	if err != nil {
		slog.Error("error scheduling task on same day", slog.Any("error", err))
	}
}

// complete: this function calls the awx set config function
func SetConfig(schedule entities.Schedule) {
	slog.Debug("setting schedule", slog.Any("schedule", schedule))
}
