package model

import "time"

type CalendarEventStatus string

const (
	CalendarEventStatusPublic    CalendarEventStatus = "public"
	CalendarEventStatusPrivate   CalendarEventStatus = "private"
	CalendarEventStatusInherited CalendarEventStatus = "inherited"
)

type CalendarStatus string

const (
	CalendarStatusPublic  CalendarStatus = "public"
	CalendarStatusPrivate CalendarStatus = "private"
)

// CalendarEvent represents an event in the calendar
// @Description Calendar event
type CalendarEvent struct {
	ID          uint64              `json:"id,string"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	StartTime   string              `json:"start_time"`
	EndTime     string              `json:"end_time"`
	Status      CalendarEventStatus `json:"status"`
}

// Calendar represents a calendar
// @Description Calendar
type Calendar struct {
	ID     uint64         `json:"id,string"`
	UserID uint64         `json:"user_id,string"`
	Name   string         `json:"name"`
	Color  string         `json:"color"`
	Status CalendarStatus `json:"status"`
	SyncedAt time.Time      `json:"sync_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GoogleCalendar represents a calendar from Google Calendar API
// @Description Google Calendar information
type GoogleCalendar struct {
	Kind                 string                              `json:"kind" example:"calendar#calendarListEntry"`
	ETag                 string                              `json:"etag" example:"\"00000000000000000000\""`
	ID                   string                              `json:"id" example:"primary"`
	Summary              string                              `json:"summary" example:"My Calendar"`
	Description          string                              `json:"description,omitempty" example:"Personal calendar"`
	Location             string                              `json:"location,omitempty" example:"Mountain View, CA"`
	TimeZone             string                              `json:"timeZone" example:"America/Los_Angeles"`
	SummaryOverride      string                              `json:"summaryOverride,omitempty" example:"Custom Summary"`
	ColorID              string                              `json:"colorId" example:"1"`
	BackgroundColor      string                              `json:"backgroundColor" example:"#9c27b0"`
	ForegroundColor      string                              `json:"foregroundColor" example:"#ffffff"`
	Hidden               bool                                `json:"hidden" example:"false"`
	Selected             bool                                `json:"selected" example:"true"`
	AccessRole           string                              `json:"accessRole" example:"owner"`
	Primary              bool                                `json:"primary" example:"true"`
	Deleted              bool                                `json:"deleted" example:"false"`
	ConferenceProperties *GoogleCalendarConferenceProperties `json:"conferenceProperties,omitempty"`
}

// GoogleCalendarConferenceProperties represents conference properties for a Google Calendar
// @Description Google Calendar conference properties
type GoogleCalendarConferenceProperties struct {
	AllowedConferenceSolutionTypes []string `json:"allowedConferenceSolutionTypes" example:"[\"hangoutsMeet\"]"`
}

// CalendarListResponse represents the response for calendar list endpoint
// @Description Calendar list response
type CalendarListResponse struct {
	Success   bool              `json:"success" example:"true"`
	Message   string            `json:"message" example:"Calendars retrieved successfully"`
	Calendars []*GoogleCalendar `json:"calendars"`
}
