package repository

import (
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
