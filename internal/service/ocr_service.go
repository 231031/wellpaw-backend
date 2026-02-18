package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

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
			"protein":  {Type: genai.TypeNumber, Description: "Regular food: crude protein percent. Supplement/serving-based product: protein grams per recommended daily amount. -1 if not found."},
			"fat":      {Type: genai.TypeNumber, Description: "Regular food: crude fat percent. Supplement/serving-based product: fat grams per recommended daily amount. -1 if not found."},
			"moisture": {Type: genai.TypeNumber, Description: "Moisture percent when available. -1 if not found."},
			"energy":   {Type: genai.TypeNumber, Description: "Regular food: energy in kcal/100g (or per recommended amount if explicitly provided). Supplement/serving-based product: kcal per recommended daily amount. -1 if not found."},
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

	mimeType := http.DetectContentType(imgData)
	all := strings.Split(mimeType, "/")
	if len(all) < 2 {
		return &model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid image type",
		}
	}
	if all[0] != "image" {
		return &model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid image type",
		}
	}
	mimeType = all[1]

	instruction := `Extract nutrition from one pet product label (one product, one type).
					Detect type: complete food, treat, topper, or supplement.
					If supplement:
					- Use feeding directions and calculate values for one recommended unit (for example 1 tsp, scoop, tablet, capsule, mL).
					- protein and fat: grams per one recommended unit.
					- energy: kcal per one recommended unit.
					- If protein/fat are given as percent, convert with:
					nutrient_grams = (percent / 100) * grams_of_recommended_unit.
					- Use grams_of_recommended_unit only from explicit label data. You may infer it only from explicit paired values (for example kcal/kg + kcal/tsp). If still unknown, return -1 for dependent fields.
					If complete food/treat/topper:
					- protein, fat, moisture = crude percentages.
					- energy = kcal/100g (convert from kcal/kg if needed).
					Use only guaranteed analysis, nutrition facts, and feeding guide values. Ignore marketing text.`

	prompt := []genai.Part{
		genai.ImageData(mimeType, imgData),
		genai.Text(instruction),
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
		candidate := resp.Candidates[0]
		if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
			return &model.HTTPResponse{
				Status:  fiber.StatusInternalServerError,
				Message: "failed to extract text from image",
			}
		}

		part := candidate.Content.Parts[0]
		txt, ok := part.(genai.Text)
		if !ok {
			return &model.HTTPResponse{
				Status:  fiber.StatusInternalServerError,
				Message: "failed to extract text from image",
			}
		}

		err := json.Unmarshal([]byte(txt), &result)
		if err != nil || result == nil {
			return &model.HTTPResponse{
				Status:  fiber.StatusInternalServerError,
				Message: "failed to extract text from image",
			}
		}
	}

	return &model.HTTPResponse{
		Status:  fiber.StatusOK,
		Message: "Text extracted successfully",
		Data:    result,
	}
}
