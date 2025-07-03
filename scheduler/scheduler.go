package scheduler

import (
	"github.com/fatemehkarimi/chronos_bot/api"
	"github.com/fatemehkarimi/chronos_bot/entities"
	"github.com/fatemehkarimi/chronos_bot/pkg/utils"
	"github.com/fatemehkarimi/chronos_bot/repository"
	"log/slog"
	"time"
)

type Scheduler interface {
	LaunchSchedulesInRange(calendar entities.Calendar, startDayTime entities.DayTime, endDayTime entities.DayTime)
	OnNewSchedule(schedule entities.Schedule)
}

type DBScheduler struct {
	repo       repository.Repository
	api        api.Api
	logChannel string
}

func NewScheduler(DB repository.Repository, api api.Api, logChannel string) Scheduler {
	return DBScheduler{DB, api, logChannel}
}

func (s DBScheduler) LaunchSchedulesInRange(
	calendar entities.Calendar, startDayTime entities.DayTime, endDayTime entities.DayTime,
) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	day := now.Day()

	schedules, err := s.repo.GetScheduleByTime(entities.GeorgianCalendar, year, month, day, startDayTime, endDayTime)

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
	if utils.ShouldRunToday(schedule) {
		go s.ScheduleAndNotify(schedule)
	}
}

func (s DBScheduler) ScheduleAndNotify(schedule entities.Schedule) {
	taskDayTime := entities.DayTime{Hour: schedule.Hour, Minute: schedule.Minute}
	task := func() error {
		SetConfig(schedule)
		s.api.SendMessage(s.logChannel, utils.ScheduleToText(schedule), nil, nil)
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
