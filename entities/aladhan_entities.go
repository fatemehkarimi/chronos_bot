package entities

type AladhanDateResponse struct {
	Code   int                `json:"code"`
	Status string             `json:"status"`
	Data   DateConversionData `json:"data"`
}

type DateConversionData struct {
	Hijri     HijriDate     `json:"hijri"`
	Gregorian GregorianDate `json:"gregorian"`
}

type HijriDate struct {
	Date             string      `json:"date"`
	Format           string      `json:"format"`
	Day              string      `json:"day"`
	Weekday          Weekday     `json:"weekday"`
	Month            HijriMonth  `json:"month"`
	Year             string      `json:"year"`
	Designation      Designation `json:"designation"`
	Holidays         []string    `json:"holidays"`
	AdjustedHolidays []string    `json:"adjustedHolidays"`
	Method           string      `json:"method"`
}

type GregorianDate struct {
	Date          string           `json:"date"`
	Format        string           `json:"format"`
	Day           string           `json:"day"`
	Weekday       GregorianWeekday `json:"weekday"`
	Month         GregorianMonth   `json:"month"`
	Year          string           `json:"year"`
	Designation   Designation      `json:"designation"`
	LunarSighting bool             `json:"lunarSighting"`
}

type Weekday struct {
	En string `json:"en"`
	Ar string `json:"ar"`
}

type GregorianWeekday struct {
	En string `json:"en"`
}

type HijriMonth struct {
	Number int    `json:"number"`
	En     string `json:"en"`
	Ar     string `json:"ar"`
	Days   int    `json:"days"`
}

type GregorianMonth struct {
	Number int    `json:"number"`
	En     string `json:"en"`
}

type Designation struct {
	Abbreviated string `json:"abbreviated"`
	Expanded    string `json:"expanded"`
}
