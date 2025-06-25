package calendar

type SaveCalendarRequest struct {
	Link string `json:"link" binding:"required"`
	Year int64  `json:"year"  binding:"required"`
}

type GetCalendarsResponse struct {
	Calendars []Calendar `json:"calendars"`
}

type DelCalendarRequest struct {
	Year int64 `json:"year"  binding:"required"`
}

type Calendar struct {
	Link string `json:"link"  binding:"required"`
	Year int64  `json:"year"  binding:"required"`
}
