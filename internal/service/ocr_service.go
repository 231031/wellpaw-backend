package service

import (
	"context"
	"encoding/json"
	"io"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
)

type OcrService interface {
	ProcessOcrRequest(ctx context.Context, file io.Reader) *model.HTTPResponse
}

type ocrService struct {
	geminiClient *genai.Client
}

func NewOcrService(geminiClient *genai.Client) OcrService {
	return &ocrService{geminiClient: geminiClient}
}

func (s *ocrService) ProcessOcrRequest(ctx context.Context, file io.Reader) *model.HTTPResponse {
	if s.geminiClient == nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "extraction service is not ready",
		}
	}

	geminiModel := s.geminiClient.GenerativeModel("gemini-2.5-flash")

	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"protein":  {Type: genai.TypeNumber, Description: "Crude Protein percentage. -1 if not found."},
			"fat":      {Type: genai.TypeNumber, Description: "Crude Fat percentage. -1 if not found."},
			"moisture": {Type: genai.TypeNumber, Description: "Moisture percentage. -1 if not found."},
			"energy":   {Type: genai.TypeNumber, Description: "Metabolizable Energy in kcal/100g. -1 if not found."},
		},
		Required: []string{"protein", "fat", "moisture", "energy"},
	}

	geminiModel.ResponseMIMEType = "application/json"
	geminiModel.ResponseSchema = schema

	imgData, err := io.ReadAll(file)
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to read image file",
		}
	}

	prompt := []genai.Part{
		genai.ImageData("jpeg", imgData),
		genai.Text("Analyze this pet food label. Extract the guaranteed analysis. Use -1 for missing values."),
	}

	resp, err := geminiModel.GenerateContent(ctx, prompt...)
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusInternalServerError,
			Message: "failed to extract text from image",
		}
	}

	var result *model.PetFoodAnalysisResponse
	if len(resp.Candidates) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if txt, ok := part.(genai.Text); ok {
			err := json.Unmarshal([]byte(txt), &result)
			if err != nil || result == nil {
				return &model.HTTPResponse{
					Status:  fiber.StatusInternalServerError,
					Message: "failed to extract text from image",
				}
			}
		}
	}

	return &model.HTTPResponse{
		Status:  fiber.StatusOK,
		Message: "Text extracted successfully",
		Data:    result,
	}
}
