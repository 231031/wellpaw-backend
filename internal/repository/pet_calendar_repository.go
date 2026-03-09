package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"gorm.io/gorm"
)

type PetCalendarRepository interface {
	CreatePetCalendar(ctx context.Context, userID uint, petCalendar *model.PetCalendar, petIDs []uint) error
	GetPetCalendarsByPetIDAndUserID(ctx context.Context, petID uint, userID uint) ([]model.PetActivityCalendar, error)
	GetPetCalendarsByUserID(ctx context.Context, userID uint) ([]model.PetCalendar, error)
	GetCurrentMonthCalendarTypeCountByUserID(ctx context.Context, userID uint, monthStart time.Time, nextMonthStart time.Time) (map[model.CalendarType]int64, error)
}

type petCalendarRepository struct {
	db *gorm.DB
}

func NewPetCalendarRepository(db *gorm.DB) PetCalendarRepository {
	return &petCalendarRepository{
		db: db,
	}
}

func (r *petCalendarRepository) CreatePetCalendar(ctx context.Context, userID uint, petCalendar *model.PetCalendar, petIDs []uint) error {
	var activityEvents []model.PetActivityCalendar

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var authorizedPetIDs []uint
		if err := tx.Model(&model.Pet{}).
			Where("user_id = ? AND id IN ?", userID, petIDs).
			Pluck("id", &authorizedPetIDs).Error; err != nil {
			return fmt.Errorf("failed to validate pet ids before creating pet calendar : %w", err)
		}

		if len(authorizedPetIDs) != len(petIDs) {
			return gorm.ErrRecordNotFound
		}

		if err := tx.Create(petCalendar).Error; err != nil {
			return fmt.Errorf("failed to create pet calendar : %w", err)
		}

		activityEvents = make([]model.PetActivityCalendar, 0, len(authorizedPetIDs))
		for _, petID := range authorizedPetIDs {
			activityEvents = append(activityEvents, model.PetActivityCalendar{
				PetID:         petID,
				PetCalendarID: petCalendar.ID,
			})
		}

		if err := tx.Create(&activityEvents).Error; err != nil {
			return fmt.Errorf("failed to create pet activity calendar : %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	petCalendar.ActivityEvents = activityEvents
	return nil
}

func (r *petCalendarRepository) GetPetCalendarsByPetIDAndUserID(ctx context.Context, petID uint, userID uint) ([]model.PetActivityCalendar, error) {
	var activityEvents []model.PetActivityCalendar

	if err := r.db.WithContext(ctx).
		Model(&model.PetActivityCalendar{}).
		Joins("JOIN pets ON pets.id = pet_activity_calendars.pet_id AND pets.deleted_at IS NULL").
		Joins("JOIN pet_calendars ON pet_calendars.id = pet_activity_calendars.pet_calendar_id AND pet_calendars.deleted_at IS NULL").
		Where("pet_activity_calendars.pet_id = ? AND pets.user_id = ?", petID, userID).
		Preload("PetCalendar").
		Preload("Pet").
		Order("pet_calendars.start_datetime ASC, pet_activity_calendars.id ASC").
		Find(&activityEvents).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet calendars by pet id : %w", err)
	}

	return activityEvents, nil
}

func (r *petCalendarRepository) GetPetCalendarsByUserID(ctx context.Context, userID uint) ([]model.PetCalendar, error) {
	var petCalendars []model.PetCalendar

	if err := r.db.WithContext(ctx).
		Model(&model.PetCalendar{}).
		Joins("JOIN pet_activity_calendars ON pet_activity_calendars.pet_calendar_id = pet_calendars.id").
		Joins("JOIN pets ON pets.id = pet_activity_calendars.pet_id AND pets.deleted_at IS NULL").
		Where("pets.user_id = ?", userID).
		Distinct("pet_calendars.*").
		Preload("ActivityEvents", func(db *gorm.DB) *gorm.DB {
			return db.
				Joins("JOIN pets ON pets.id = pet_activity_calendars.pet_id AND pets.deleted_at IS NULL").
				Where("pets.user_id = ?", userID).
				Preload("Pet").
				Order("pet_activity_calendars.id ASC")
		}).
		Order("pet_calendars.start_datetime ASC, pet_calendars.id ASC").
		Find(&petCalendars).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet calendars by user id : %w", err)
	}

	return petCalendars, nil
}

func (r *petCalendarRepository) GetCurrentMonthCalendarTypeCountByUserID(ctx context.Context, userID uint, monthStart time.Time, nextMonthStart time.Time) (map[model.CalendarType]int64, error) {
	type calendarTypeCountRow struct {
		Type  model.CalendarType `gorm:"column:type"`
		Times int64              `gorm:"column:times"`
	}

	var rows []calendarTypeCountRow
	if err := r.db.WithContext(ctx).
		Model(&model.PetCalendar{}).
		Select("pet_calendars.type AS type, COUNT(DISTINCT pet_calendars.id) AS times").
		Joins("JOIN pet_activity_calendars ON pet_activity_calendars.pet_calendar_id = pet_calendars.id").
		Joins("JOIN pets ON pets.id = pet_activity_calendars.pet_id AND pets.deleted_at IS NULL").
		Where("pets.user_id = ?", userID).
		Where("pet_calendars.start_datetime >= ? AND pet_calendars.start_datetime < ?", monthStart, nextMonthStart).
		Group("pet_calendars.type").
		Order("pet_calendars.type ASC").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to get current month pet calendar summary by user id : %w", err)
	}

	summaryByType := make(map[model.CalendarType]int64, len(rows))
	for _, row := range rows {
		summaryByType[row.Type] = row.Times
	}

	return summaryByType, nil
}
