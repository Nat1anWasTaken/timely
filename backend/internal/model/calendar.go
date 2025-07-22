package model

import (
	"time"

	"gorm.io/gorm"
)

type CalendarEventVisibility string

const (
	CalendarEventVisibilityPublic    CalendarEventVisibility = "public"
	CalendarEventVisibilityPrivate   CalendarEventVisibility = "private"
	CalendarEventVisibilityInherited CalendarEventVisibility = "inherited"
)

type CalendarVisibility string

const (
	CalendarVisibilityPublic  CalendarVisibility = "public"
	CalendarVisibilityPrivate CalendarVisibility = "private"
)

type CalendarSource string

const (
	SourceGoogle CalendarSource = "google"
	SourceISC    CalendarSource = "isc"
)

// CalendarEvent represents an event in the calendar
// @Description Calendar event
type CalendarEvent struct {
	ID          uint64                  `json:"id,string"`          // Google Event ID
	SourceID    string                  `json:"source_id"`          // Source calendar ID
	CalendarID  uint64                  `json:"calendar_id,string"` // calendar ID
	Title       string                  `json:"title"`              // Event title (summary)
	Start       time.Time               `json:"start"`              // ISO8601 datetime or date (for all-day)
	End         time.Time               `json:"end"`                // ISO8601 datetime or date
	AllDay      bool                    `json:"allDay"`             // True if it's an all-day event
	EventColor  string                  `json:"event_color"`        // Optional display color
	Location    string                  `json:"location"`           // Optional event location
	Description string                  `json:"description"`        // Optional description
	Visibility  CalendarEventVisibility `json:"visibility"`         // public / private / default
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	DeletedAt   gorm.DeletedAt          `json:"-" gorm:"index"`
}

// Calendar represents a calendar
// @Description Calendar
type Calendar struct {
	ID            uint64             `json:"id,string" gorm:"primaryKey"`
	UserID        uint64             `json:"user_id,string" gorm:"index"`
	SourceID      *string            `json:"source_id"`
	Source        CalendarSource     `json:"source"`
	Summary       string             `json:"summary"`
	TimeZone      string             `json:"time_zone"`
	Description   *string            `json:"description,omitempty"`
	EventNickname *string            `json:"event_nickname,omitempty"`
	EventColor    *string            `json:"event_color,omitempty"`
	Visibility    CalendarVisibility `json:"visibility"`
	SyncedAt      time.Time          `json:"synced_at"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	DeletedAt     gorm.DeletedAt     `json:"-" gorm:"index"`
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

// GoogleCalendarEvent represents an event from Google Calendar API
// @Description Google Calendar event information
type GoogleCalendarEvent struct {
	Kind        string                      `json:"kind" example:"calendar#event"`
	ETag        string                      `json:"etag" example:"\"00000000000000000000\""`
	ID          string                      `json:"id" example:"event_id_123"`
	Status      string                      `json:"status" example:"confirmed"`
	HTMLLink    string                      `json:"htmlLink" example:"https://www.google.com/calendar/event?eid=..."`
	Created     string                      `json:"created" example:"2024-01-01T00:00:00.000Z"`
	Updated     string                      `json:"updated" example:"2024-01-01T00:00:00.000Z"`
	Summary     string                      `json:"summary" example:"Meeting with team"`
	Description string                      `json:"description" example:"Weekly team meeting"`
	Location    string                      `json:"location" example:"Conference Room A"`
	ColorID     string                      `json:"colorId" example:"1"`
	Creator     *GoogleCalendarEventActor   `json:"creator,omitempty"`
	Organizer   *GoogleCalendarEventActor   `json:"organizer,omitempty"`
	Start       *GoogleCalendarEventTime    `json:"start"`
	End         *GoogleCalendarEventTime    `json:"end"`
	Visibility  string                      `json:"visibility" example:"default"`
	Attendees   []*GoogleCalendarEventActor `json:"attendees,omitempty"`
}

// GoogleCalendarEventActor represents a creator, organizer, or attendee of an event
// @Description Google Calendar event actor information
type GoogleCalendarEventActor struct {
	ID          string `json:"id" example:"user@example.com"`
	Email       string `json:"email" example:"user@example.com"`
	DisplayName string `json:"displayName" example:"John Doe"`
	Self        bool   `json:"self" example:"true"`
}

// GoogleCalendarEventTime represents the start or end time of an event
// @Description Google Calendar event time information
type GoogleCalendarEventTime struct {
	Date     string `json:"date" example:"2024-01-01"`               // For all-day events
	DateTime string `json:"dateTime" example:"2024-01-01T10:00:00Z"` // For timed events
	TimeZone string `json:"timeZone" example:"America/Los_Angeles"`
}

// GoogleCalendarEventsResponse represents the response from Google Calendar events API
// @Description Google Calendar events response
type GoogleCalendarEventsResponse struct {
	Kind     string                 `json:"kind" example:"calendar#events"`
	ETag     string                 `json:"etag" example:"\"00000000000000000000\""`
	Summary  string                 `json:"summary" example:"My Calendar"`
	Updated  string                 `json:"updated" example:"2024-01-01T00:00:00.000Z"`
	TimeZone string                 `json:"timeZone" example:"America/Los_Angeles"`
	Items    []*GoogleCalendarEvent `json:"items"`
}

// CalendarEventsRequest represents the request for getting calendar events
// @Description Calendar events request with time range
type CalendarEventsRequest struct {
	StartTime string `json:"start_time" validate:"required" example:"2024-01-01T00:00:00Z"`
	EndTime   string `json:"end_time" validate:"required" example:"2024-01-31T23:59:59Z"`
}

// CalendarWithEvents represents a calendar with its events
// @Description Calendar with events
type CalendarWithEvents struct {
	*Calendar
	Events []*CalendarEvent `json:"events"`
}

// CalendarEventsResponse represents the response for calendar events endpoint
// @Description Calendar events response
type CalendarEventsResponse struct {
	Success   bool                  `json:"success" example:"true"`
	Message   string                `json:"message" example:"Calendar events retrieved successfully"`
	Calendars []*CalendarWithEvents `json:"calendars"`
}
