package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type PetService interface {
	CreateNewPet(ctx context.Context, pet *model.PetPayload) *model.HTTPResponse
	GetPetByID(ctx context.Context, petID uint) *model.HTTPResponse
	GetPetsByUserID(ctx context.Context, userID uint) *model.HTTPResponse
	GetPetAnalysisByPetID(ctx context.Context, petID uint) *model.HTTPResponse
	UpdatePetInfo(ctx context.Context, petInfo *model.Pet) *model.HTTPResponse
	UpdatePetDetail(ctx context.Context, petDetail *model.PetDetail) *model.HTTPResponse
	SoftDeletePet(ctx context.Context, userID uint, petID uint) *model.HTTPResponse
}

type petService struct {
	calculationService    CalculationService
	petRepo               repository.PetRepository
	petFoodPlanRepo       repository.PetFoodPlanRepository
	freeValidationService FreeTierUsageValidationService
}

func NewPetService(calculationService CalculationService, petRepo repository.PetRepository, petFoodPlanRepo repository.PetFoodPlanRepository, freeTierUsageValidationService FreeTierUsageValidationService) PetService {
	return &petService{
		calculationService:    calculationService,
		petRepo:               petRepo,
		petFoodPlanRepo:       petFoodPlanRepo,
		freeValidationService: freeTierUsageValidationService,
	}
}

// not test
func (s *petService) GetAgeRangeFromBirthDate(petType model.PetType, birthDate time.Time) model.AgeType {
	ageMonth := time.Since(birthDate).Hours() / 24 / 30
	switch petType {
	case model.CAT:
		if ageMonth <= 12 {
			return model.JUNIOR
		} else if ageMonth >= 84 {
			return model.SENIOR
		} else {
			return model.ADULT
		}
	case model.DOG:
		if ageMonth <= 12 {
			return model.JUNIOR
		} else if ageMonth >= 84 {
			return model.SENIOR
		} else {
			return model.ADULT
		}
	}
	return model.ADULT
}

func (s *petService) checkValidPetInfoDetails(petInfo *model.Pet, petDetail *model.PetDetail) *model.HTTPResponse {
	if *petInfo.SexType == model.MALE && petDetail.Gestation != nil && *petDetail.Gestation {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "pet detail update failed because pet is male",
		}
	}
	if *petInfo.SexType == model.MALE && petDetail.Lactation != nil && *petDetail.Lactation {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "pet detail update failed because pet is male",
		}
	}
	return nil
}

func (s *petService) CreateNewPet(ctx context.Context, pet *model.PetPayload) *model.HTTPResponse {
	// tier, freeUsage, resp := s.freeValidationService.CheckValidUsageByUserID(ctx, pet.PetInfo.UserID, model.PROFILE)
	// if resp != nil {
	// 	return resp
	// }

	petInfo := pet.PetInfo
	petDetail := pet.PetDetail

	resp := s.checkValidPetInfoDetails(petInfo, petDetail)
	if resp != nil {
		return resp
	}

	petDetail.AgeRange = s.GetAgeRangeFromBirthDate(*petInfo.Type, petInfo.BirthDate)
	petDetail.Energy = s.calculationService.CalMerEnergyRequirement(petDetail, *petInfo.Type)
	petDetail.Protein, petDetail.Fat = s.calculationService.CalNutritientRequirement(petDetail.Energy, petDetail, *petInfo.Type)

	err := s.petRepo.CreateNewPet(ctx, petInfo, petDetail)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "new pet",
		}
	}

	// if tier != nil && *tier == model.FREE {
	// 	freeUsage.ProfileFree += 1
	// 	s.freeValidationService.UpdateFreeTierUsage(ctx, petInfo.UserID, freeUsage)
	// }

	petDetail.CreatedAt = utils.ConvertTimeToThaiTimezone(petDetail.CreatedAt)
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data: map[string]interface{}{
			"pet_info":   petInfo,
			"pet_detail": petDetail,
		},
	}
}

func (s *petService) GetPetsByUserID(ctx context.Context, userID uint) *model.HTTPResponse {
	pets, err := s.petRepo.GetPetsByUserID(ctx, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pets",
		}
	}

	for idx := range pets {
		s.convertPetCreatedAtToThaiTimezone(&pets[idx])
		if len(pets[idx].PetDetails) > 0 {
			s.convertPetDetailCreatedAtToThaiTimezone(&pets[idx].PetDetails[0])
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pets": pets,
		},
	}
}

func (s *petService) GetPetByID(ctx context.Context, petID uint) *model.HTTPResponse {
	pet, err := s.petRepo.GetPetInfoByID(ctx, petID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "not found pet",
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet",
		}
	}

	petDetail := model.PetDetail{}
	if len(pet.PetDetails) > 0 {
		petDetail = pet.PetDetails[0]
	}

	s.convertPetCreatedAtToThaiTimezone(pet)
	s.convertPetDetailCreatedAtToThaiTimezone(&petDetail)
	if len(pet.PetDetails) > 0 {
		s.convertPetDetailCreatedAtToThaiTimezone(&pet.PetDetails[0])
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pet_info":   pet,
			"pet_detail": petDetail,
		},
	}
}

func (s *petService) GetPetAnalysisByPetID(ctx context.Context, petID uint) *model.HTTPResponse {
	pet, err := s.petRepo.GetPetAnalysisByID(ctx, petID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet not found",
			}
		}
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet analysis",
		}
	}

	planUsageHistories, err := s.petFoodPlanRepo.GetPlanUsageHistoryByPetID(ctx, petID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusInternalServerError,
				Message: utils.FailedToGetMsg + "pet food plan history",
			}
		}
		planUsageHistories = []model.PetFoodPlanHistory{}
	}

	avgWeight := s.calculationService.CalAvgPercentWeightChangePerMonth(pet.MonthlyNutritionTWA, pet.PetDetails[0].BCS, *pet.Type, pet.PetDetails[0].AgeRange)
	for i := 1; i < len(pet.MonthlyNutritionTWA); i++ {
		prev := pet.MonthlyNutritionTWA[i-1].Weight
		curr := pet.MonthlyNutritionTWA[i].Weight
		if prev <= 0 {
			continue
		}

		pet.MonthlyNutritionTWA[i].PercentWeightChange = ((curr - prev) / prev) * 100
	}

	for _, p := range planUsageHistories {
		if p.PetFoodPlanTotal.PetFoodPlan.Unit == model.CUP {
			if len(p.PetFoodPlanTotal.PetFoodPlanDetails) > 0 {
				for idx, d := range p.PetFoodPlanTotal.PetFoodPlanDetails {
					cup := s.calculationService.CalculateGramsToCup(d)
					p.PetFoodPlanTotal.PetFoodPlanDetails[idx].Cup = cup
				}
			}
		}
	}

	if len(pet.PetDetails) > 0 {
		s.convertPetDetailCreatedAtToThaiTimezone(&pet.PetDetails[0])
	}
	for idx := range planUsageHistories {
		s.convertPetFoodPlanHistoryCreatedAtToThaiTimezone(&planUsageHistories[idx])
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data: map[string]interface{}{
			"pet":                                 pet,
			"pet_food_plan_histories":             planUsageHistories,
			"avg_percent_weight_change_per_month": avgWeight,
		},
	}
}

func (s *petService) UpdatePetInfo(ctx context.Context, petInfo *model.Pet) *model.HTTPResponse {
	if err := s.petRepo.UpdatePetInfo(ctx, petInfo); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet info",
		}
	}

	petInfo, err := s.petRepo.GetPetInfoByID(ctx, petInfo.ID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "pet information is updated, but failed to get new data",
		}
	}

	s.convertPetCreatedAtToThaiTimezone(petInfo)
	if len(petInfo.PetDetails) > 0 {
		s.convertPetDetailCreatedAtToThaiTimezone(&petInfo.PetDetails[0])
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   petInfo,
	}
}

func (s *petService) mapUpdatePetDetailNotTrigger(plan *model.PetFoodPlan) (*model.PetFoodPlanTotal, []*model.PetFoodPlanDetail, *model.HTTPResponse) {
	if len(plan.PetFoodPlanTotals) <= 0 {
		applogger.LogInfo(fmt.Sprintf("plan id %d doesn't have pet food total", plan.ID), serviceLog)
		return nil, nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet",
		}
	}

	foodPlanDetails := []*model.PetFoodPlanDetail{}
	foodPlanTotal := &model.PetFoodPlanTotal{
		PetFoodPlanID:      plan.ID,
		TotalEnergyIntake:  plan.PetFoodPlanTotals[0].TotalEnergyIntake,
		TotalProteinIntake: plan.PetFoodPlanTotals[0].TotalProteinIntake,
		TotalFatIntake:     plan.PetFoodPlanTotals[0].TotalFatIntake,
	}

	for _, d := range plan.PetFoodPlanTotals[0].PetFoodPlanDetails {
		foodPlanDetails = append(foodPlanDetails, &model.PetFoodPlanDetail{
			FoodPetFoodPlanID: d.FoodPetFoodPlanID,
			Amount:            d.Amount,
			EnergyIntake:      d.EnergyIntake,
			ProteinIntake:     d.ProteinIntake,
			FatIntake:         d.FatIntake,
		})
	}
	return foodPlanTotal, foodPlanDetails, nil
}

func (s *petService) UpdatePetDetail(ctx context.Context, petDetail *model.PetDetail) *model.HTTPResponse {
	pet, err := s.petRepo.GetPetInfoByID(ctx, petDetail.PetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet",
		}
	}

	if len(pet.PetDetails) > 0 {
		if *pet.PetDetails[0].Neutered && petDetail.Neutered != nil && !*petDetail.Neutered {
			return &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: "pet detail update failed because pet is already neutered",
			}
		}
		resp := s.checkValidPetInfoDetails(pet, petDetail)
		if resp != nil {
			return resp
		}
	}

	petDetail.AgeRange = s.GetAgeRangeFromBirthDate(*pet.Type, pet.BirthDate)
	petDetail.Energy = s.calculationService.CalMerEnergyRequirement(petDetail, *pet.Type)
	petDetail.Protein, petDetail.Fat = s.calculationService.CalNutritientRequirement(petDetail.Energy, petDetail, *pet.Type)

	latestPlanID, foodsInActivePlan, err := s.petFoodPlanRepo.GetFoodsInLastestActivePlanByPetID(ctx, pet.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := s.petRepo.UpdatePetDetails(ctx, petDetail); err != nil {
				return &model.HTTPResponse{
					Status:  http.StatusInternalServerError,
					Message: utils.FailedToUpdateMsg + "pet detail",
				}
			}

			s.convertPetCreatedAtToThaiTimezone(pet)
			s.convertPetDetailCreatedAtToThaiTimezone(petDetail)
			pet.PetDetails[0] = *petDetail
			if len(pet.PetDetails) > 0 {
				s.convertPetDetailCreatedAtToThaiTimezone(&pet.PetDetails[0])
			}

			return &model.HTTPResponse{
				Status: http.StatusOK,
				Data: map[string]interface{}{
					"pet_info":   pet,
					"pet_detail": petDetail,
				},
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet detail",
		}
	}

	if len(foodsInActivePlan) <= 0 {
		applogger.LogInfo(fmt.Sprintf("plan id : %d not found any food in plan", latestPlanID), serviceLog)
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet detail",
		}
	}

	var foods []model.Food
	var resp *model.HTTPResponse
	foodPlanDetails := []*model.PetFoodPlanDetail{}
	foodPlanTotal := &model.PetFoodPlanTotal{}

	if foodsInActivePlan[0].PetFoodPlan.SelfDefine {
		foodPlanTotal, foodPlanDetails, resp = s.mapUpdatePetDetailNotTrigger(foodsInActivePlan[0].PetFoodPlan)
		if resp != nil {
			return resp
		}
	} else {
		supAmount := map[uint]float64{}
		for _, fp := range foodsInActivePlan {
			foods = append(foods, *fp.Food)
			if *fp.Food.Type == model.SUPPLEMENTS {
				supAmount[fp.FoodID] = fp.PetFoodPlanDetails[0].Amount
			}
		}
		foodPlanDetails = s.calculationService.CalFeedingAmountPerDay(petDetail, foods, supAmount)
		for idx := range foodsInActivePlan {
			foodPlanDetails[idx].FoodPetFoodPlanID = foodsInActivePlan[idx].ID
		}

		foodPlanTotal = s.calculationService.CalTotalIntakeFoodPlan(foodPlanDetails)
		foodPlanTotal.PetFoodPlanID = latestPlanID
	}

	if err := s.petRepo.UpdatePetDetailsAndPlan(ctx, pet.ID, petDetail, foodPlanTotal, foodPlanDetails); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet detail",
		}
	}

	activePlanDetail, err := s.petFoodPlanRepo.GetLastestActivePlanDetailByPet(ctx, pet.ID, petDetail.ID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "successfully updated pet detail and plan but " + utils.FailedToGetMsg + "lastest pet food plan",
		}
	}

	if activePlanDetail.Unit == model.CUP {
		s.calculationService.ConvertGramsToCupInPlan(activePlanDetail)
	}

	pet.PetDetails[0] = *petDetail
	s.convertPetCreatedAtToThaiTimezone(pet)
	s.convertPetDetailCreatedAtToThaiTimezone(petDetail)
	s.convertPetFoodPlanCreatedAtToThaiTimezone(activePlanDetail)

	if len(pet.PetDetails) > 0 {
		s.convertPetDetailCreatedAtToThaiTimezone(&pet.PetDetails[0])
	}

	responseData := map[string]interface{}{
		"pet_info":      pet,
		"pet_detail":    petDetail,
		"pet_food_plan": activePlanDetail,
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   responseData,
	}
}

func (s *petService) convertPetCreatedAtToThaiTimezone(pet *model.Pet) {
	if pet == nil {
		return
	}

	pet.CreatedAt = utils.ConvertTimeToThaiTimezone(pet.CreatedAt)
}

func (s *petService) convertPetDetailCreatedAtToThaiTimezone(petDetail *model.PetDetail) {
	if petDetail == nil {
		return
	}

	petDetail.CreatedAt = utils.ConvertTimeToThaiTimezone(petDetail.CreatedAt)
}

func (s *petService) convertPetFoodPlanCreatedAtToThaiTimezone(petFoodPlan *model.PetFoodPlan) {
	if petFoodPlan == nil {
		return
	}

	petFoodPlan.CreatedAt = utils.ConvertTimeToThaiTimezone(petFoodPlan.CreatedAt)
}

func (s *petService) convertPetFoodPlanHistoryCreatedAtToThaiTimezone(history *model.PetFoodPlanHistory) {
	if history == nil {
		return
	}

	history.CreatedAt = utils.ConvertTimeToThaiTimezone(history.CreatedAt)
	history.PlanUsageEndDate = utils.ConvertTimeToThaiTimezone(history.PlanUsageEndDate)
	if history.PetFoodPlanTotal == nil {
		return
	}

	history.PetFoodPlanTotal.CreatedAt = utils.ConvertTimeToThaiTimezone(history.PetFoodPlanTotal.CreatedAt)
	if history.PetFoodPlanTotal.PetFoodPlan == nil {
		return
	}

	history.PetFoodPlanTotal.PetFoodPlan.CreatedAt = utils.ConvertTimeToThaiTimezone(history.PetFoodPlanTotal.PetFoodPlan.CreatedAt)
}

func (s *petService) SoftDeletePet(ctx context.Context, userID uint, petID uint) *model.HTTPResponse {
	if err := s.petRepo.SoftDeletePetByIDAndUserID(ctx, petID, userID); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to delete pet",
		}
	}

	return &model.HTTPResponse{
		Status:  http.StatusOK,
		Message: "pet deleted successfully",
	}
}
