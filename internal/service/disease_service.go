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
	PredictPetSkinDisease(ctx context.Context, userID uint, payload *model.PredictPetSkinDiseasePayload) *model.HTTPResponse
	GetPetSkinImagesByUserID(ctx context.Context, userID uint) *model.HTTPResponse
	GetPetSkinImagesByPetID(ctx context.Context, userID uint, petID uint) *model.HTTPResponse
	LabeledPetSkinDisease(ctx context.Context, userID uint, payload *model.LabeledPetSkinDiseasePayload) *model.HTTPResponse
	mapPetSkinClassToDiseaseType(petType model.PetType, classIndex int) (model.DiseaseType, bool)
}

type diseaseService struct {
	modelService          ModelService
	petRepo               repository.PetRepository
	petSkinImageRepo      repository.PetSkinImageRepository
	freeValidationService FreeTierUsageValidationService
}

func NewDiseaseService(modelService ModelService, petRepo repository.PetRepository, petSkinImageRepo repository.PetSkinImageRepository, freeTierUsageValidationService FreeTierUsageValidationService) DiseaseService {
	return &diseaseService{
		modelService:          modelService,
		petRepo:               petRepo,
		petSkinImageRepo:      petSkinImageRepo,
		freeValidationService: freeTierUsageValidationService,
	}
}

func (s *diseaseService) PredictPetSkinDisease(ctx context.Context, userID uint, payload *model.PredictPetSkinDiseasePayload) *model.HTTPResponse {
	tier, freeUsage, resp := s.freeValidationService.CheckValidUsageByUserID(ctx, userID, model.DISEASE)
	if resp != nil {
		return resp
	}

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

	// check base64 format that valid or not if not return 400 status invalid image file
	if err := utils.ValidateBase64Image(payload.Image); err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid image file",
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

	if tier != nil && *tier == model.FREE {
		freeUsage.DiseaseFree += 1
		s.freeValidationService.UpdateFreeTierUsage(ctx, userID, freeUsage)
	}

	petSkinImage.Disease = petSkinImage.Predicted.String()
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data:   petSkinImage,
	}
}

func (s *diseaseService) GetPetSkinImagesByUserID(ctx context.Context, userID uint) *model.HTTPResponse {
	petSkinImages, err := s.petSkinImageRepo.GetPetSkinImagesByUserID(ctx, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet skin images",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   s.mapPetSkinImagesDisease(petSkinImages),
	}
}

func (s *diseaseService) GetPetSkinImagesByPetID(ctx context.Context, userID uint, petID uint) *model.HTTPResponse {
	petSkinImages, err := s.petSkinImageRepo.GetPetSkinImagesByPetIDAndUserID(ctx, petID, userID)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToGetMsg + "pet skin images",
		}
	}

	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   s.mapPetSkinImagesDisease(petSkinImages),
	}
}

func (s *diseaseService) LabeledPetSkinDisease(ctx context.Context, userID uint, payload *model.LabeledPetSkinDiseasePayload) *model.HTTPResponse {
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

	if !s.isValidPetSkinDiseaseType(*pet.Type, *payload.Labeled) {
		return &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid skin disease type " + pet.Type.String(),
		}
	}

	if err := s.petSkinImageRepo.UpdateLabeledPetSkinDiseaseByID(ctx, payload.PetSkinImageID, payload.PetID, userID, *payload.Labeled, payload.ImageEvidence); err != nil {
		if errors.Is(err, utils.ErrNoRowsUpdated) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet skin image" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: utils.FailedToUpdateMsg + "pet skin image label",
		}
	}

	updatedPetSkinImage, err := s.petSkinImageRepo.GetPetSkinImageByID(ctx, payload.PetSkinImageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.HTTPResponse{
				Status:  http.StatusNotFound,
				Message: "pet skin image" + utils.NotFoundMsg,
			}
		}

		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "pet skin image labeled, but " + utils.FailedToGetMsg + "updated pet skin image",
		}
	}

	updatedPetSkinImage.Disease = updatedPetSkinImage.Predicted.String()
	return &model.HTTPResponse{
		Status: http.StatusOK,
		Data:   updatedPetSkinImage,
	}
}

func (s *diseaseService) mapPetSkinImagesDisease(petSkinImages []model.PetSkinImage) []model.PetSkinImage {
	if len(petSkinImages) == 0 {
		return []model.PetSkinImage{}
	}

	for idx := range petSkinImages {
		petSkinImages[idx].Disease = petSkinImages[idx].Predicted.String()
	}

	return petSkinImages
}

func (s *diseaseService) isValidPetSkinDiseaseType(petType model.PetType, diseaseType model.DiseaseType) bool {
	switch petType {
	case model.CAT:
		return diseaseType == model.HEALTHY ||
			diseaseType == model.OTHER ||
			diseaseType == model.RINGWORM ||
			diseaseType == model.SCABIES
	case model.DOG:
		return diseaseType == model.BACTERIAL_DERMATOSIS ||
			diseaseType == model.DEMODICOSIS ||
			diseaseType == model.HEALTHY ||
			diseaseType == model.OTHER ||
			diseaseType == model.RINGWORM
	default:
		return false
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
