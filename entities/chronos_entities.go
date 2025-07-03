package entities

type FeatureFlag struct {
	Name     string
	OwnerId  int
	UnixTime int64
}

type Schedule struct {
	ScheduleId      int
	FeatureFlagName string
	Value           string
	UsersList       string
	Calendar        CalendarTime
	UnixTime        int64
}

type State int

const (
	_ State = iota
	StartState

	// add feature flag
	AddFeatureFlagState

	// add scheduler
	ChooseFeatureFlagState
	ChooseCalendarTypeState
	GetScheduleState
	GetValueState
	GetUserListState
)

type UserState struct {
	StateName State

	// for scheduler state
	Schedule *Schedule
}

type CalendarType int

const (
	_ CalendarType = iota
	KhorshidiCalendarType
	GeorgianCalendarType
	QamariCalendarType
)

type CalendarTime struct {
	Type   CalendarType
	Year   int
	Month  int
	Day    int
	Hour   int
	Minute int
}
