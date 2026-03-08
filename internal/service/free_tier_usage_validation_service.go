package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/applogger"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type FreeTierUsageValdiationService interface {
	CheckValidUsageByUserID(ctx context.Context, userID uint, usageFeature model.UsageFeatureType) (*model.TierType, *model.FreeTierUsage, *model.HTTPResponse)
	CheckFreeTierDiseaseUsage(diseaseFree int) *model.HTTPResponse
	CheckFreeTierProfileUsage(profileFree int) *model.HTTPResponse
	CheckFreeTierFoodPlanUsage(foodPlanFree int) *model.HTTPResponse
	CheckFreeTierFoodUsage(foodFree int) *model.HTTPResponse
}

type freeTierUsageValdiationService struct {
	userRepo          repository.UserRepository
	freeTierUsageRepo repository.FreeTierUsageRepository
}

func NewFreeTierUsageValdiation(userRepository repository.UserRepository, freeTierUsageRepo repository.FreeTierUsageRepository) FreeTierUsageValdiationService {
	return &freeTierUsageValdiationService{
		userRepo:          userRepository,
		freeTierUsageRepo: freeTierUsageRepo,
	}
}

func (s *freeTierUsageValdiationService) CheckValidUsageByUserID(ctx context.Context, userID uint, usageFeature model.UsageFeatureType) (*model.TierType, *model.FreeTierUsage, *model.HTTPResponse) {
	tier, status, err := s.userRepo.GetSubscriptionDetail(ctx, userID)
	if err != nil {
		return nil, nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.ErrFailToCheckFreeTierUsage.Error(),
		}
	}

	if tier != nil && *tier != model.FREE {
		if status != nil && *status == model.ACTIVESUB {
			return tier, nil, nil
		} else {
			return tier, nil, &model.HTTPResponse{
				Status:  http.StatusBadRequest,
				Message: fmt.Sprintf("invalid subscription status: %s, cannot use this feature", status.String()),
			}
		}
	}

	if status != nil && *status != model.ACTIVESUB {
		return tier, nil, &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("invalid subscription status: %s for free tial, subscribe first", status.String()),
		}
	}

	freeUsage, err := s.getFreeTierUsageInfoByUserID(ctx, userID)
	if err != nil {
		return tier, nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.ErrFailToCheckFreeTierUsage.Error(),
		}
	}

	switch usageFeature {
	case model.DISEASE:
		return tier, freeUsage, s.CheckFreeTierDiseaseUsage(freeUsage.DiseaseFree)
	case model.FOODPLAN:
		return tier, freeUsage, s.CheckFreeTierFoodPlanUsage(freeUsage.FoodPlanFree)
	case model.FOOD:
		return tier, freeUsage, s.CheckFreeTierFoodUsage(freeUsage.FoodFree)
	case model.PROFILE:
		return tier, freeUsage, s.CheckFreeTierProfileUsage(freeUsage.ProfileFree)
	default:
		return tier, freeUsage, &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid usage feature",
		}
	}
}

func (s *freeTierUsageValdiationService) getFreeTierUsageInfoByUserID(ctx context.Context, userID uint) (*model.FreeTierUsage, error) {
	freeUsage, err := s.freeTierUsageRepo.GetDeleteFreeTierUsage(ctx, userID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			user, err := s.userRepo.GetSubscriptionDetailFromDB(ctx, userID)
			if err != nil {
				return nil, utils.ErrFailToCheckFreeTierUsage
			}

			freeUsage = &model.FreeTierUsage{
				FoodFree:     user.FoodFree,
				FoodPlanFree: user.FoodPlanFree,
				ProfileFree:  user.ProfileFree,
				DiseaseFree:  user.DiseaseFree,
			}

			err = s.freeTierUsageRepo.SetFreeTierUsage(ctx, userID, freeUsage)
			if err != nil {
				applogger.LogError(fmt.Sprintf("failed to set free tier usage in redis : %v", err), serviceLog)
			}

			return freeUsage, nil
		}

		applogger.LogError(fmt.Sprintf("failed to get free tier usage in redis : %v", err), serviceLog)
		return nil, utils.ErrFailToCheckFreeTierUsage
	}

	return freeUsage, nil
}

func (s *freeTierUsageValdiationService) CheckFreeTierDiseaseUsage(diseaseFree int) *model.HTTPResponse {
	if diseaseFree >= 1 {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "skin disease prediction usage reached limation in free tial",
		}
	}
	return nil
}

func (s *freeTierUsageValdiationService) CheckFreeTierProfileUsage(profileFree int) *model.HTTPResponse {
	if profileFree >= 1 {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "pet profile reached limation in free tial",
		}
	}
	return nil
}

func (s *freeTierUsageValdiationService) CheckFreeTierFoodPlanUsage(foodPlanFree int) *model.HTTPResponse {
	if foodPlanFree >= 1 {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "food plan creation usage reached limation in free tial",
		}
	}
	return nil
}

func (s *freeTierUsageValdiationService) CheckFreeTierFoodUsage(foodFree int) *model.HTTPResponse {
	if foodFree >= 3 {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "food creation usage reached limation in free tial",
		}
	}
	return nil
}
