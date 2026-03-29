package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type PetRepository interface {
	CreateNewPet(ctx context.Context, pet *model.Pet, petDetails *model.PetDetail) error
	UpdatePetInfo(ctx context.Context, pet *model.Pet) error
	UpdatePetDetails(ctx context.Context, petDetails *model.PetDetail) error
	UpdatePetDetailsAndPlan(ctx context.Context, petID uint, petDetails *model.PetDetail, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error
	GetPetsLatestOneMonthPetDetail(ctx context.Context, latestID uint) ([]model.Pet, error)
	GetPetsByUserID(ctx context.Context, id uint) ([]model.Pet, error)
	GetPetAnalysisByID(ctx context.Context, id uint) (*model.Pet, error)
	GetPetInfoByID(ctx context.Context, id uint) (*model.Pet, error)
	GetLatestPetDetailByPetID(ctx context.Context, petID uint) (*model.PetDetail, error)
	GetPetDetialsFromPetID(ctx context.Context, petID uint) ([]model.PetDetail, error)
	SoftDeletePetByIDAndUserID(ctx context.Context, petID uint, userID uint) error
}

type petRepository struct {
	db *gorm.DB
}

func NewPetRepository(db *gorm.DB) PetRepository {
	return &petRepository{
		db: db,
	}
}

func (r *petRepository) CreateNewPet(ctx context.Context, pet *model.Pet, petDetails *model.PetDetail) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(pet).Error; err != nil {
			return fmt.Errorf("failed to create new pet : %w", err)
		}

		petDetails.PetID = pet.ID
		if err := tx.Create(petDetails).Error; err != nil {
			return fmt.Errorf("failed to create pet details : %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

func (r *petRepository) UpdatePetInfo(ctx context.Context, pet *model.Pet) error {
	result := r.db.WithContext(ctx).Model(&model.Pet{}).
		Where("id = ?", pet.ID).
		Updates(map[string]interface{}{
			"image_path": pet.ImagePath,
			"name":       pet.Name,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to update pet info : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}

func (r *petRepository) UpdatePetDetails(ctx context.Context, petDetails *model.PetDetail) error {
	if err := r.db.Create(petDetails).Error; err != nil {
		return err
	}

	return nil
}

func (r *petRepository) UpdatePetDetailsAndPlan(ctx context.Context, petID uint, petDetails *model.PetDetail, foodPlanTotal *model.PetFoodPlanTotal, foodPlanDetails []*model.PetFoodPlanDetail) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(petDetails).Error; err != nil {
			return fmt.Errorf("failed to update pet details : %w", err)
		}

		foodPlanTotal.PetDetailID = petDetails.ID
		if err := tx.Create(foodPlanTotal).Error; err != nil {
			return err
		}

		for idx := range foodPlanDetails {
			foodPlanDetails[idx].PetFoodPlanTotalID = foodPlanTotal.ID
		}
		if err := tx.Create(foodPlanDetails).Error; err != nil {
			return fmt.Errorf("failed to update food plan detail : %w", err)
		}

		planHistory := &model.PetFoodPlanHistory{
			PetID:              petID,
			PetFoodPlanTotalID: foodPlanTotal.ID,
		}
		if err := tx.Create(planHistory).Error; err != nil {
			return fmt.Errorf("failed to update new history of food : %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *petRepository) GetPetInfoByID(ctx context.Context, id uint) (*model.Pet, error) {
	var pet *model.Pet
	err := r.db.WithContext(ctx).Preload("PetDetails", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC, id DESC").Limit(1)
	}).First(&pet, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by id : %w", err)
	}

	return pet, nil
}

func (r *petRepository) GetPetsLatestOneMonthPetDetail(ctx context.Context, latestID uint) ([]model.Pet, error) {
	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	pets := []model.Pet{}
	err := r.db.WithContext(ctx).
		Joins("JOIN users ON pets.user_id = users.id").
		Where("pets.id > ? AND users.noti_update_pet = ?", latestID, true).
		Preload("User").
		Order("pets.id ASC").
		Limit(100).
		Find(&pets).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pet info: %w", err)
	}

	if len(pets) == 0 {
		return pets, nil
	}

	petIDs := make([]uint, 0, len(pets))
	petIndexByID := make(map[uint]int, len(pets))
	for i, p := range pets {
		petIDs = append(petIDs, p.ID)
		petIndexByID[p.ID] = i
	}

	detailsQuery := `
		WITH recent_details AS (
			SELECT
				id,
				pet_id,
				weight,
				activity_level,
				age_range,
				bcs,
				lactation,
				gestation,
				gestation_start_date,
				neutered,
				energy,
				protein,
				fat,
				created_at
			FROM pet_details
			WHERE pet_id IN ?
				AND created_at >= ?
		),
		previous_details AS (
			SELECT
				id,
				pet_id,
				weight,
				activity_level,
				age_range,
				bcs,
				lactation,
				gestation,
				gestation_start_date,
				neutered,
				energy,
				protein,
				fat,
				created_at
			FROM (
				SELECT
					id,
					pet_id,
					weight,
					activity_level,
					age_range,
					bcs,
					lactation,
					gestation,
					gestation_start_date,
					neutered,
					energy,
					protein,
					fat,
					created_at,
					ROW_NUMBER() OVER (
						PARTITION BY pet_id
						ORDER BY created_at DESC, id DESC
					) AS rn
				FROM pet_details
				WHERE pet_id IN ?
					AND created_at < ?
			) ranked
			WHERE rn = 1
		),
		relevant_details AS (
			SELECT * FROM recent_details
			UNION ALL
			SELECT * FROM previous_details
		)
		SELECT
			id,
			pet_id,
			weight,
			activity_level,
			age_range,
			bcs,
			lactation,
			gestation,
			gestation_start_date,
			neutered,
			energy,
			protein,
			fat,
			created_at
		FROM relevant_details
		ORDER BY pet_id ASC, created_at DESC, id DESC;
	`

	details := []model.PetDetail{}
	if err := r.db.WithContext(ctx).Raw(detailsQuery, petIDs, oneMonthAgo, petIDs, oneMonthAgo).Scan(&details).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet details for monthly reminder: %w", err)
	}

	for _, detail := range details {
		idx, ok := petIndexByID[detail.PetID]
		if !ok {
			continue
		}
		pets[idx].PetDetails = append(pets[idx].PetDetails, detail)
	}

	return pets, nil
}

func (r *petRepository) GetLatestPetDetailByPetID(ctx context.Context, petID uint) (*model.PetDetail, error) {
	var petDetail model.PetDetail
	if err := r.db.WithContext(ctx).
		Where("pet_id = ?", petID).
		Order("created_at DESC, id DESC").
		First(&petDetail).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest pet detail by pet id : %w", err)
	}

	return &petDetail, nil
}

func (r *petRepository) GetPetDetialsFromPetID(ctx context.Context, petID uint) ([]model.PetDetail, error) {
	currentDate := time.Now()
	oneYearAgo := currentDate.AddDate(-1, 0, 0)

	query := `
		WITH in_window_details AS (
			SELECT *
			FROM pet_details
			WHERE pet_id = ?
				AND created_at >= ?
				AND created_at <= ?
		),
		last_before_window AS (
			SELECT *
			FROM pet_details
			WHERE pet_id = ?
				AND created_at < ?
			ORDER BY created_at DESC, id DESC
			LIMIT 1
		),
		relevant_details AS (
			SELECT * FROM in_window_details
			UNION ALL
			SELECT * FROM last_before_window
		)
		SELECT *
		FROM relevant_details
		ORDER BY created_at ASC, id ASC;
	`

	var petDetails []model.PetDetail
	if err := r.db.WithContext(ctx).Raw(
		query,
		petID,
		oneYearAgo,
		currentDate,
		petID,
		oneYearAgo,
	).Scan(&petDetails).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet detail by pet id : %w", err)
	}

	return petDetails, nil
}

func (r *petRepository) GetPetAnalysisByID(ctx context.Context, id uint) (*model.Pet, error) {
	loc, _ := time.LoadLocation("Asia/Bangkok")
	nowThai := time.Now().In(loc)
	oneYearAgo := nowThai.AddDate(-1, 0, 0)

	var pet model.Pet
	err := r.db.WithContext(ctx).Preload("PetDetails", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(1)
	}).First(&pet, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get pet by id : %w", err)
	}

	// Time-Weighted Average (TWA) by month over the last year.
	monthlyTWA := []model.PetMonthlyNutritionTWA{}
	twaQuery := `
		WITH in_window_history AS (
			SELECT
				h.id,
				h.created_at,
				h.pet_food_plan_total_id
			FROM pets p
			JOIN pet_food_plan_histories h ON h.pet_id = p.id
			WHERE p.id = ?
				AND p.deleted_at IS NULL
				AND h.created_at >= ?
				AND h.created_at <= ?
		),
		last_before_window AS (
			SELECT
				h.id,
				h.created_at,
				h.pet_food_plan_total_id
			FROM pets p
			JOIN pet_food_plan_histories h ON h.pet_id = p.id
			WHERE p.id = ?
				AND p.deleted_at IS NULL
				AND h.created_at < ?
			ORDER BY h.created_at DESC, h.id DESC
			LIMIT 1
		),
		relevant_history AS (
			SELECT * FROM in_window_history
			UNION ALL
			SELECT * FROM last_before_window
		),
		history_intervals AS (
			SELECT
				h.id AS history_id,
				h.created_at AS interval_start,
				COALESCE(
					LEAD(h.created_at) OVER (ORDER BY h.created_at, h.id),
					?
				) AS interval_end,
				t.total_energy_intake,
				t.total_protein_intake,
				t.total_fat_intake,
				d.energy,
				d.protein,
				d.fat,
				d.weight,
				d.activity_level,
				d.bcs
			FROM relevant_history h
			JOIN pet_food_plan_totals t ON t.id = h.pet_food_plan_total_id
			JOIN pet_details d ON d.id = t.pet_detail_id
		),
		bounded_intervals AS (
			SELECT
				history_id,
				interval_start AS source_created_at,
				GREATEST(interval_start, ?) AS effective_start,
				LEAST(interval_end, ?) AS effective_end,
				total_energy_intake,
				total_protein_intake,
				total_fat_intake,
				energy,
				protein,
				fat,
				weight,
				activity_level,
				bcs
			FROM history_intervals
		),
		split_by_month AS (
			SELECT
				gs.month_start,
				b.history_id,
				b.source_created_at,
				GREATEST(b.effective_start, gs.month_start) AS segment_start,
				LEAST(b.effective_end, gs.month_start + INTERVAL '1 month') AS segment_end,
				b.total_energy_intake,
				b.total_protein_intake,
				b.total_fat_intake,
				b.energy,
				b.protein,
				b.fat,
				b.weight,
				b.activity_level,
				b.bcs
			FROM bounded_intervals b
			JOIN LATERAL generate_series(
				date_trunc('month', b.effective_start AT TIME ZONE 'Asia/Bangkok'),
				date_trunc('month', b.effective_end AT TIME ZONE 'Asia/Bangkok'),
				INTERVAL '1 month'
			) AS gs(month_start) ON TRUE
			WHERE b.effective_end > b.effective_start
		),
		weighted_data AS (
			SELECT
				EXTRACT(YEAR FROM month_start)::int AS year,
				EXTRACT(MONTH FROM month_start)::int AS month,
				history_id,
				source_created_at,
				total_energy_intake,
				total_protein_intake,
				total_fat_intake,
				energy,
				protein,
				fat,
				weight,
				activity_level,
				bcs,
				EXTRACT(EPOCH FROM (segment_end - segment_start)) AS weight_seconds
			FROM split_by_month
			WHERE segment_end > segment_start
		)
		SELECT
			year,
			month,
			SUM(total_energy_intake * weight_seconds) / SUM(weight_seconds) AS total_energy_intake,
			SUM(total_protein_intake * weight_seconds) / SUM(weight_seconds) AS total_protein_intake,
			SUM(total_fat_intake * weight_seconds) / SUM(weight_seconds) AS total_fat_intake,
			SUM(energy * weight_seconds) / SUM(weight_seconds) AS energy,
			SUM(protein * weight_seconds) / SUM(weight_seconds) AS protein,
			SUM(fat * weight_seconds) / SUM(weight_seconds) AS fat,
			(ARRAY_AGG(weight ORDER BY source_created_at DESC, history_id DESC))[1] AS weight,
			(ARRAY_AGG(activity_level ORDER BY source_created_at DESC, history_id DESC))[1]::int AS activity_level,
			(ARRAY_AGG(bcs ORDER BY source_created_at DESC, history_id DESC))[1]::int AS bcs
		FROM weighted_data
		GROUP BY year, month
		ORDER BY year, month;
	`
	if err := r.db.WithContext(ctx).Raw(
		twaQuery,
		// cte1
		id,
		oneYearAgo, // lower bound of analysis window
		nowThai,    // upper bound of analysis window
		// cte2
		id,
		oneYearAgo, // pick latest row before analysis window
		// cte4
		nowThai, // fallback interval_end for latest history row
		// cte5
		oneYearAgo, // lower bound of analysis window
		nowThai,    // upper bound of analysis window
	).Scan(&monthlyTWA).Error; err != nil {
		return nil, fmt.Errorf("failed to get monthly TWA by pet id : %w", err)
	}

	pet.MonthlyNutritionTWA = monthlyTWA
	return &pet, nil
}

func (r *petRepository) GetPetsByUserID(ctx context.Context, id uint) ([]model.Pet, error) {
	var pets []model.Pet
	if err := r.db.WithContext(ctx).Where("user_id = ?", id).Find(&pets).Error; err != nil {
		return nil, fmt.Errorf("failed to get pet by user id : %w", err)
	}

	if len(pets) == 0 {
		return pets, nil
	}

	petIDs := make([]uint, 0, len(pets))
	petIndexByID := make(map[uint]int, len(pets))
	for idx, pet := range pets {
		petIDs = append(petIDs, pet.ID)
		petIndexByID[pet.ID] = idx
	}

	latestDetailQuery := `
		SELECT
			id,
			pet_id,
			weight,
			activity_level,
			age_range,
			bcs,
			lactation,
			gestation,
			gestation_start_date,
			neutered,
			energy,
			protein,
			fat,
			created_at
		FROM (
			SELECT
				id,
				pet_id,
				weight,
				activity_level,
				age_range,
				bcs,
				lactation,
				gestation,
				gestation_start_date,
				neutered,
				energy,
				protein,
				fat,
				created_at,
				ROW_NUMBER() OVER (
					PARTITION BY pet_id
					ORDER BY created_at DESC, id DESC
				) AS rn
			FROM pet_details
			WHERE pet_id IN ?
		) ranked
		WHERE rn = 1;
	`

	latestDetails := []model.PetDetail{}
	if err := r.db.WithContext(ctx).Raw(latestDetailQuery, petIDs).Scan(&latestDetails).Error; err != nil {
		return nil, fmt.Errorf("failed to get latest pet details by user id : %w", err)
	}

	for _, detail := range latestDetails {
		idx, ok := petIndexByID[detail.PetID]
		if !ok {
			continue
		}
		pets[idx].PetDetails = []model.PetDetail{detail}
	}

	return pets, nil
}

func (r *petRepository) SoftDeletePetByIDAndUserID(ctx context.Context, petID uint, userID uint) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", petID, userID).
		Delete(&model.Pet{})
	if result.Error != nil {
		return fmt.Errorf("failed to soft delete pet : %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return utils.ErrNoRowsUpdated
	}

	return nil
}
