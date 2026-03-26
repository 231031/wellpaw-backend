package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/repository"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/google/uuid"
	gcpstorage "google.golang.org/api/storage/v1"
	"gorm.io/gorm"
)

type DiseaseService interface {
	validateAndPredictDisease(ctx context.Context, imageBase string, petType model.PetType) (model.DiseaseType, int, *model.HTTPResponse)
	PredictPetSkinDiseaseUnknown(ctx context.Context, userID uint, payload *model.PredictPetSkinDiseaseUnknownPayload) *model.HTTPResponse
	PredictPetSkinDisease(ctx context.Context, userID uint, payload *model.PredictPetSkinDiseasePayload) *model.HTTPResponse
	GetPetSkinImagesByUserID(ctx context.Context, userID uint) *model.HTTPResponse
	GetPetSkinImagesByPetID(ctx context.Context, userID uint, petID uint) *model.HTTPResponse
	LabeledPetSkinDisease(ctx context.Context, userID uint, payload *model.LabeledPetSkinDiseasePayload) *model.HTTPResponse
	UploadSkinImage(userID uint, petID uint, predicted model.DiseaseType, imageBase64 string) (string, error)
	mapPetSkinClassToDiseaseType(petType model.PetType, classIndex int) (model.DiseaseType, bool)
}

type diseaseService struct {
	modelService          ModelService
	petRepo               repository.PetRepository
	petSkinImageRepo      repository.PetSkinImageRepository
	freeValidationService FreeTierUsageValidationService
	firebaseStorage       *model.FirebaseStorage
}

func NewDiseaseService(modelService ModelService, petRepo repository.PetRepository, petSkinImageRepo repository.PetSkinImageRepository, freeTierUsageValidationService FreeTierUsageValidationService, firebaseStorage *model.FirebaseStorage) DiseaseService {
	return &diseaseService{
		modelService:          modelService,
		petRepo:               petRepo,
		petSkinImageRepo:      petSkinImageRepo,
		freeValidationService: freeTierUsageValidationService,
		firebaseStorage:       firebaseStorage,
	}
}

func (s *diseaseService) convertConfidentToRangeConfident(confident int) int {
	switch {
	case confident >= 80:
		return 5
	case confident >= 60:
		return 4
	case confident >= 40:
		return 3
	case confident >= 20:
		return 2
	default:
		return 1
	}
}

func (s *diseaseService) validateAndPredictDisease(ctx context.Context, imageBase string, petType model.PetType) (model.DiseaseType, int, *model.HTTPResponse) {
	// check base64 format that valid or not if not return 400 status invalid image file
	if err := utils.ValidateBase64Image(imageBase); err != nil {
		return -1, 0, &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "invalid image file",
		}
	}

	modelResp, httpResp := s.modelService.PredictPetSkinDisease(ctx, imageBase, petType)
	if httpResp != nil {
		return -1, 0, httpResp
	}

	predicted, ok := s.mapPetSkinClassToDiseaseType(petType, modelResp.ClassIndex)
	if !ok {
		return -1, 0, &model.HTTPResponse{
			Status:  http.StatusBadGateway,
			Message: "invalid prediction from model api",
		}
	}

	confident := int(math.Round(modelResp.Probability * 100))
	if confident < 0 {
		confident = 0
	}
	if confident > 100 {
		confident = 100
	}

	return predicted, confident, nil
}

func (s *diseaseService) PredictPetSkinDiseaseUnknown(ctx context.Context, userID uint, payload *model.PredictPetSkinDiseaseUnknownPayload) *model.HTTPResponse {
	// tier, freeUsage, resp := s.freeValidationService.CheckValidUsageByUserID(ctx, userID, model.DISEASE)
	// if resp != nil {
	// 	return resp
	// }

	predicted, confident, resp := s.validateAndPredictDisease(ctx, payload.Image, *payload.PetType)
	if resp != nil {
		return resp
	}
	petSkinImage := &model.PetSkinImage{
		ImagePath: "",
		Predicted: predicted,
		Labeled:   -1,
		Confident: confident,
	}

	// if tier != nil && *tier == model.FREE {
	// 	freeUsage.DiseaseFree += 1
	// 	s.freeValidationService.UpdateFreeTierUsage(ctx, userID, freeUsage)
	// }

	petSkinImage.Disease = petSkinImage.Predicted.String()
	petSkinImage.Confident = s.convertConfidentToRangeConfident(petSkinImage.Confident)
	return &model.HTTPResponse{
		Status: http.StatusCreated,
		Data:   petSkinImage,
	}
}

func (s *diseaseService) PredictPetSkinDisease(ctx context.Context, userID uint, payload *model.PredictPetSkinDiseasePayload) *model.HTTPResponse {
	// tier, freeUsage, resp := s.freeValidationService.CheckValidUsageByUserID(ctx, userID, model.DISEASE)
	// if resp != nil {
	// 	return resp
	// }

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

	predicted, confident, resp := s.validateAndPredictDisease(ctx, payload.Image, *pet.Type)
	if resp != nil {
		return resp
	}

	imagePath, err := s.UploadSkinImage(userID, pet.ID, predicted, payload.Image)
	if err != nil {
		return &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to upload pet skin image",
		}
	}

	petSkinImage := &model.PetSkinImage{
		PetID:     pet.ID,
		ImagePath: imagePath,
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

	// if tier != nil && *tier == model.FREE {
	// 	freeUsage.DiseaseFree += 1
	// 	s.freeValidationService.UpdateFreeTierUsage(ctx, userID, freeUsage)
	// }

	petSkinImage.CreatedAt = utils.ConvertTimeToThaiTimezone(petSkinImage.CreatedAt)
	petSkinImage.Disease = petSkinImage.Predicted.String()
	petSkinImage.Confident = s.convertConfidentToRangeConfident(petSkinImage.Confident)
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
	updatedPetSkinImage.Confident = s.convertConfidentToRangeConfident(updatedPetSkinImage.Confident)
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
		petSkinImages[idx].CreatedAt = utils.ConvertTimeToThaiTimezone(petSkinImages[idx].CreatedAt)
		petSkinImages[idx].Disease = petSkinImages[idx].Predicted.String()
		petSkinImages[idx].Confident = s.convertConfidentToRangeConfident(petSkinImages[idx].Confident)
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

func (s *diseaseService) UploadSkinImage(userID uint, petID uint, predicted model.DiseaseType, imageBase64 string) (string, error) {
	decodedImage, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		decodedImage, err = base64.RawStdEncoding.DecodeString(imageBase64)
		if err != nil {
			return "", fmt.Errorf("invalid base64 image: %w", err)
		}
	}

	contentType := http.DetectContentType(decodedImage)
	ext := utils.DetectContentType(contentType)

	predictedPath := strings.ToLower(strings.ReplaceAll(predicted.String(), " ", "_"))
	objectName := fmt.Sprintf("disease/%s/%d/%d/%s%s", predictedPath, userID, petID, uuid.NewString(), ext)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	_, err = s.firebaseStorage.Objects.Insert(s.firebaseStorage.BucketName, &gcpstorage.Object{
		Name:        objectName,
		ContentType: contentType,
	}).Media(bytes.NewReader(decodedImage)).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload image to firebase storage: %w", err)
	}

	encoded := url.PathEscape(objectName)
	url := fmt.Sprintf(
		"https://firebasestorage.googleapis.com/v0/b/%s/o/%s?alt=media",
		s.firebaseStorage.BucketName,
		encoded,
	)

	return url, nil
}
