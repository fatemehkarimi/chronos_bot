package entities

type FeatureFlag struct {
	Name     string
	OwnerId  int
	UnixTime int64
}

type State int

const (
	_ State = iota
	StartState

	// add feature flag
	AddFeatureFlagState

	// add schedule
	ChooseFeatureFlagState
	ChooseCalendarTypeState
)

type UserState struct {
	StateName State

	// for schedule state
	SelectedFeatureFlag  *FeatureFlag
	SelectedCalendarType *Calendar
}

type Calendar int

const (
	_ Calendar = iota
	KhorshidiCalendar
	GeorgianCalendar
	QamariCalendar
)
