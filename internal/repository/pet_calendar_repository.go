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
	GetActiveActivityCalendar(ctx context.Context, lastID uint) ([]model.PetCalendar, error)
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

// 1
// 7,12
// 22,24
// 16,17,19
// 11
func (r *petCalendarRepository) GetActiveActivityCalendar(ctx context.Context, lastID uint) ([]model.PetCalendar, error) {
	query := `
		WITH calendar_candidates AS (
			SELECT DISTINCT
				pc.id,
				pc.start_datetime,
				pc.end_date,
				pc.frequently
			FROM pet_calendars pc
			JOIN pet_activity_calendars pac ON pac.pet_calendar_id = pc.id
			JOIN pets p ON p.id = pac.pet_id AND p.deleted_at IS NULL
			JOIN users u ON u.id = p.user_id AND u.deleted_at IS NULL
			WHERE
				u.noti_calendars = TRUE
				AND u.device_token <> ''
				AND pc.deleted_at IS null
				AND pc.end_date >= CURRENT_DATE
				AND CURRENT_DATE >= (pc.start_datetime)::date
		),
		due_now AS (
			SELECT c.id
			FROM calendar_candidates c
			where
				(
					c.frequently = ?
					AND NOW() - INTERVAL '1 minutes' <= c.start_datetime
					AND c.start_datetime <= NOW() + INTERVAL '1 minutes'
				)
				OR
				(
					c.frequently = ?
					AND ABS(EXTRACT(EPOCH FROM (c.start_datetime::time - NOW()::time))) < 120
				)
				OR
				(
					c.frequently = ?
					AND TO_CHAR(CURRENT_DATE, 'Day') = TO_CHAR(c.start_datetime, 'Day')
					AND ABS(EXTRACT(EPOCH FROM (c.start_datetime::time - NOW()::time))) < 120
				)
				OR
				(
					c.frequently = ?
					AND DATE_PART('day', CURRENT_DATE) = DATE_PART('day', c.start_datetime)
					AND ABS(EXTRACT(EPOCH FROM (c.start_datetime::time - NOW()::time))) < 120
				)
				OR
				(
					c.frequently = ?
					AND DATE_PART('day', CURRENT_DATE) = DATE_PART('day', c.start_datetime)
					AND EXTRACT(MONTH FROM CURRENT_DATE) = EXTRACT(MONTH FROM c.start_datetime)
					AND ABS(EXTRACT(EPOCH FROM (c.start_datetime::time - NOW()::time))) < 120

				)
		),

		target_ids AS (
			SELECT id
			FROM due_now
			WHERE id > ?
			ORDER BY id
			LIMIT 450
		)

		SELECT id FROM target_ids;
	`

	var calendarIDs []uint
	if err := r.db.WithContext(ctx).Raw(
		query,
		model.NOT,
		model.DAY,
		model.WEEK,
		model.MONTH,
		model.YEAR,
		lastID,
	).Scan(&calendarIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get active activity calendar ids : %w", err)
	}

	if len(calendarIDs) == 0 {
		return []model.PetCalendar{}, nil
	}

	var calendars []model.PetCalendar
	if err := r.db.WithContext(ctx).
		Model(&model.PetCalendar{}).
		Where("pet_calendars.id IN ?", calendarIDs).
		Order("pet_calendars.id ASC").
		Preload("ActivityEvents", func(db *gorm.DB) *gorm.DB {
			return db.Order("pet_activity_calendars.id ASC")
		}).
		Preload("ActivityEvents.Pet", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "user_id", "name", "type", "breed", "sex_type")
		}).
		Preload("ActivityEvents.Pet.User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "device_token")
		}).
		Find(&calendars).Error; err != nil {
		return nil, fmt.Errorf("failed to get active activity calendars : %w", err)
	}

	return calendars, nil
}
