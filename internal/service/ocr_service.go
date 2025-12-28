package service

import (
	"context"
	"io"

	vision "cloud.google.com/go/vision/apiv1"
	"cloud.google.com/go/vision/v2/apiv1/visionpb"
	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/gofiber/fiber/v2"
)

type OcrService interface {
	ProcessOcrRequest(ctx context.Context, file io.Reader) *model.HTTPResponse
}

type ocrService struct {
}

func NewOcrService() OcrService {
	return &ocrService{}
}

func (s *ocrService) ProcessOcrRequest(ctx context.Context, file io.Reader) *model.HTTPResponse {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to create OCR client",
		}
	}
	defer client.Close()

	image, err := vision.NewImageFromReader(file)
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "Failed to create image from file",
		}
	}

	annotations, err := client.DetectDocumentText(ctx, image, nil)
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to detect text",
		}
	}

	if annotations == nil || annotations.Text == "" {
		return &model.HTTPResponse{
			Status:  fiber.StatusOK,
			Message: "No text found in image",
			Data:    fiber.Map{"text": ""},
		}
	}

	// extract keywords
	_, err = s.ExtractKeywords(annotations)
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "Failed to extract keywords",
		}
	}

	return &model.HTTPResponse{
		Status:  fiber.StatusOK,
		Message: "Text extracted successfully",
		Data: fiber.Map{
			"text": annotations.Text,
		},
	}
}

func (s *ocrService) ExtractKeywords(annotations *visionpb.TextAnnotation) (*model.FoodDetailResponse, error) {
	return &model.FoodDetailResponse{}, nil
}
