package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/231031/wellpaw-backend/internal/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
)

type OcrService interface {
	ProcessOcrRequest(ctx context.Context, userID uint, imageBase64 string) *model.HTTPResponse
}

type ocrService struct {
	geminiClient          *genai.Client
	freeValidationService FreeTierUsageValidationService
}

func NewOcrService(geminiClient *genai.Client, freeTierUsageValidationService FreeTierUsageValidationService) OcrService {
	return &ocrService{
		geminiClient:          geminiClient,
		freeValidationService: freeTierUsageValidationService,
	}
}

func (s *ocrService) ProcessOcrRequest(ctx context.Context, userID uint, imageBase64 string) *model.HTTPResponse {
	// _, _, usageResp := s.freeValidationService.CheckValidUsageByUserID(ctx, userID, model.FOOD)
	// if usageResp != nil {
	// 	return usageResp
	// }

	imageBase64 = normalizeBase64ImageInput(imageBase64)
	if imageBase64 == "" {
		return &model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid image file",
		}
	}
	if err := utils.ValidateBase64Image(imageBase64); err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid image file",
		}
	}

	imgData, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		imgData, err = base64.RawStdEncoding.DecodeString(imageBase64)
	}
	if err != nil {
		return &model.HTTPResponse{
			Status:  fiber.StatusBadRequest,
			Message: "invalid image file",
		}
	}

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
			"protein":  {Type: genai.TypeNumber, Description: "Regular food: crude protein percent. Supplement/serving-based product: protein grams per one serving unit. -1 if not found."},
			"fat":      {Type: genai.TypeNumber, Description: "Regular food: crude fat percent. Supplement/serving-based product: fat grams per one serving unit. -1 if not found."},
			"moisture": {Type: genai.TypeNumber, Description: "Moisture percent when available. -1 if not found."},
			"energy":   {Type: genai.TypeNumber, Description: "Regular food: energy in kcal/100g (or per recommended amount if explicitly provided). Supplement/serving-based product: kcal per one serving unit. -1 if not found."},
		},
		Required: []string{"protein", "fat", "moisture", "energy"},
	}

	geminiModel.ResponseMIMEType = "application/json"
	geminiModel.ResponseSchema = schema

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
					- Use feeding directions and calculate values for one serving unit (for example 1 tsp, scoop, tablet, capsule, mL).
					- protein and fat: grams per one serving unit.
					- energy: kcal per one serving unit.
					- If protein/fat are given as percent, convert with:
					nutrient_grams = (percent_nutritient / 100) * grams_of_serving_unit (find grams of each 1 serving unit such as  tsp, scoop, tablet, capsule, mL - consider about wet or dry).
					- Use grams_of_serving_unit only from explicit label data. You may infer it only from explicit paired values (for example kcal/kg + kcal/tsp). If still unknown, return -1 for dependent fields.
					end for If supplement:
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

func normalizeBase64ImageInput(image string) string {
	image = strings.TrimSpace(image)
	if strings.HasPrefix(image, "data:") {
		if idx := strings.Index(image, ","); idx != -1 {
			return strings.TrimSpace(image[idx+1:])
		}
	}
	return image
}
