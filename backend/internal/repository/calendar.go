package repository

import (
	"time"

	"gorm.io/gorm"

	"github.com/NathanWasTaken/timely/backend/internal/model"
)

type CalendarRepository struct {
	db *gorm.DB
}

func NewCalendarRepository(db *gorm.DB) *CalendarRepository {
	return &CalendarRepository{
		db: db,
	}
}

// Create creates a new calendar
func (r *CalendarRepository) Create(calendar *model.Calendar) error {
	return r.db.Create(calendar).Error
}

// FindByID finds a calendar by ID
func (r *CalendarRepository) FindByID(id string) (*model.Calendar, error) {
	var calendar model.Calendar
	err := r.db.Where("id = ?", id).First(&calendar).Error
	if err != nil {
		return nil, err
	}
	return &calendar, nil
}

// FindByUserID finds all calendars for a user
func (r *CalendarRepository) FindByUserID(userID uint64) ([]*model.Calendar, error) {
	var calendars []*model.Calendar
	err := r.db.Where("user_id = ?", userID).Find(&calendars).Error
	if err != nil {
		return nil, err
	}
	return calendars, nil
}

// FindByUserIDAndSourceID finds a calendar by user ID and source ID
func (r *CalendarRepository) FindByUserIDAndSourceID(userID uint64, sourceID string) (*model.Calendar, error) {
	var calendar model.Calendar
	err := r.db.Where("user_id = ? AND source_id = ?", userID, sourceID).First(&calendar).Error
	if err != nil {
		return nil, err
	}
	return &calendar, nil
}

// Update updates an existing calendar
func (r *CalendarRepository) Update(calendar *model.Calendar) error {
	return r.db.Save(calendar).Error
}

// Delete deletes a calendar
func (r *CalendarRepository) Delete(id string) error {
	return r.db.Delete(&model.Calendar{}, "id = ?", id).Error
}

// ExistsByUserIDAndSourceID checks if a calendar exists for a user with the given source ID
func (r *CalendarRepository) ExistsByUserIDAndSourceID(userID uint64, sourceID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Calendar{}).
		Where("user_id = ? AND source_id = ?", userID, sourceID).
		Count(&count).Error
	return count > 0, err
}

// CalendarEvent repository methods

// CreateEvent creates a new calendar event
func (r *CalendarRepository) CreateEvent(event *model.CalendarEvent) error {
	return r.db.Create(event).Error
}

// CreateEvents creates multiple calendar events in a batch
func (r *CalendarRepository) CreateEvents(events []*model.CalendarEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.CreateInBatches(events, 100).Error
}

// FindEventsByCalendarID finds all events for a specific calendar
func (r *CalendarRepository) FindEventsByCalendarID(calendarID uint64) ([]*model.CalendarEvent, error) {
	var events []*model.CalendarEvent
	err := r.db.Where("calendar_id = ?", calendarID).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// FindEventsByUserID finds all events for a user across all their calendars
func (r *CalendarRepository) FindEventsByUserID(userID uint64) ([]*model.CalendarEvent, error) {
	var events []*model.CalendarEvent
	err := r.db.Joins("JOIN calendars ON calendar_events.calendar_id = calendars.id").
		Where("calendars.user_id = ?", userID).
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// DeleteEventsByCalendarID deletes all events for a specific calendar
func (r *CalendarRepository) DeleteEventsByCalendarID(calendarID uint64) error {
	return r.db.Where("calendar_id = ?", calendarID).Delete(&model.CalendarEvent{}).Error
}

// ExistsEventBySourceID checks if an event exists by its source ID
func (r *CalendarRepository) ExistsEventBySourceID(sourceID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.CalendarEvent{}).Where("source_id = ?", sourceID).Count(&count).Error
	return count > 0, err
}

// FindEventsByCalendarIDsAndTimeRange finds events for specific calendars within a time range
func (r *CalendarRepository) FindEventsByCalendarIDsAndTimeRange(calendarIDs []uint64, startTime, endTime time.Time) ([]*model.CalendarEvent, error) {
	var events []*model.CalendarEvent
	err := r.db.Where("calendar_id IN ? AND start >= ? AND end <= ?", calendarIDs, startTime, endTime).
		Order("start ASC").
		Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}
