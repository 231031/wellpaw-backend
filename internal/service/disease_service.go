package service

import (
	"context"
	"errors"
	"math"
	"net/http"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"gorm.io/gorm"
)

type DiseaseService interface {
	PredictPetSkinDisease(ctx context.Context, payload *model.PredictPetSkinDiseasePayload) *model.HTTPResponse
	mapPetSkinClassToDiseaseType(petType model.PetType, classIndex int) (model.DiseaseType, bool)
}

type diseaseService struct {
	modelService     ModelService
	petRepo          repository.PetRepository
	petSkinImageRepo repository.PetSkinImageRepository
}

func NewDiseaseService(modelService ModelService, petRepo repository.PetRepository, petSkinImageRepo repository.PetSkinImageRepository) DiseaseService {
	return &diseaseService{
		modelService:     modelService,
		petRepo:          petRepo,
		petSkinImageRepo: petSkinImageRepo,
	}
}

func (s *diseaseService) PredictPetSkinDisease(ctx context.Context, payload *model.PredictPetSkinDiseasePayload) *model.HTTPResponse {
	pet, err := s.petRepo.GetPetInfoByID(ctx, payload.PetID)
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

	modelResp, httpResp := s.modelService.PredictPetSkinDisease(ctx, *pet.Type, payload)
	if httpResp != nil {
		return httpResp
	}

	predicted, ok := s.mapPetSkinClassToDiseaseType(*pet.Type, modelResp.ClassIndex)
	if !ok {
		return &model.HTTPResponse{
			Status:  http.StatusBadGateway,
			Message: "invalid class_index from model api",
		}
	}

	confident := int(math.Round(modelResp.Probability * 100))
	if confident < 0 {
		confident = 0
	}
	if confident > 100 {
		confident = 100
	}

	petSkinImage := &model.PetSkinImage{
		PetID:     pet.ID,
		ImagePath: "",
		Predicted: predicted,
		Labeled:   -1,
		Confident: confident,
	}

	if err := s.petSkinImageRepo.CreatePetSkinImage(ctx, petSkinImage); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToCreateMsg + "pet skin image",
		}
	}

	petSkinImage.Disease = petSkinImage.Predicted.String()
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data: map[string]interface{}{
			"pet_skin_image": petSkinImage,
		},
	}
}

func (s *diseaseService) mapPetSkinClassToDiseaseType(petType model.PetType, classIndex int) (model.DiseaseType, bool) {
	switch petType {
	case model.CAT:
		switch classIndex {
		case 0:
			return model.HEALTHY, true
		case 1:
			return model.OTHER, true
		case 2:
			return model.RINGWORM, true
		case 3:
			return model.SCABIES, true
		default:
			return 0, false
		}
	case model.DOG:
		switch classIndex {
		case 0:
			return model.BACTERIAL_DERMATOSIS, true
		case 1:
			return model.DEMODICOSIS, true
		case 2:
			return model.HEALTHY, true
		case 3:
			return model.OTHER, true
		case 4:
			return model.RINGWORM, true
		default:
			return 0, false
		}
	default:
		return 0, false
	}
}
