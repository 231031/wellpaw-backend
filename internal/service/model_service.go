package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/231031/wellpaw-backend/internal/model"
)

type ModelService interface {
	PredictPetSkinDisease(ctx context.Context, imageBase string, petType model.PetType) (*model.PetSkinModelResponse, *model.HTTPResponse)
}

type modelService struct {
	modelBaseAPI string
	httpClient   *http.Client
}

func NewModelService(modelBaseAPI string) ModelService {
	return &modelService{
		modelBaseAPI: strings.TrimRight(modelBaseAPI, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *modelService) PredictPetSkinDisease(ctx context.Context, imageBase string, petType model.PetType) (*model.PetSkinModelResponse, *model.HTTPResponse) {
	predictURL, ok := s.buildPredictURL(petType)
	if !ok {
		return nil, &model.HTTPResponse{
			Status:  http.StatusBadRequest,
			Message: "unsupported pet type",
		}
	}

	return s.callPredictAPI(ctx, predictURL, imageBase)
}

func (s *modelService) buildPredictURL(petType model.PetType) (string, bool) {
	switch petType {
	case model.DOG:
		return s.modelBaseAPI + "/predict/dog", true
	case model.CAT:
		return s.modelBaseAPI + "/predict/cat", true
	default:
		return "", false
	}
}

func (s *modelService) callPredictAPI(ctx context.Context, predictURL string, image string) (*model.PetSkinModelResponse, *model.HTTPResponse) {
	modelPayload := &model.PredictPetSkinModelPayload{
		Image: image,
	}

	reqBody, err := json.Marshal(modelPayload)
	if err != nil {
		return nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to encode prediction payload",
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, predictURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, &model.HTTPResponse{
			Status:  http.StatusInternalServerError,
			Message: "failed to create prediction request",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil, &model.HTTPResponse{
			Status:  http.StatusBadGateway,
			Message: "failed to call model api",
		}
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &model.HTTPResponse{
			Status:  http.StatusBadGateway,
			Message: "failed to read model api response",
		}
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &model.HTTPResponse{
			Status:  http.StatusBadGateway,
			Message: "model api request failed",
		}
	}

	var modelResp model.PetSkinModelResponse
	if err := json.Unmarshal(rawBody, &modelResp); err != nil {
		return nil, &model.HTTPResponse{
			Status:  http.StatusBadGateway,
			Message: "failed to decode model api response",
		}
	}

	return &modelResp, nil
}
