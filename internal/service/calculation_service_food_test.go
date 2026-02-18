package service

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"testing"

	"github.com/231031/wellpaw-backend/internal/model"
	"github.com/stretchr/testify/assert"
)

const (
	calFeedingCaseEnv = "CAL_FEEDING_CASES"
	feedingCasesFile  = "feeding_testcase.json"
	feedingFoodsFile  = "food_test.json"
)

type calFeedingCase struct {
	Name              string                    `json:"name"`
	PetEnergy         float64                   `json:"pet_energy"`
	FoodIndexes       []int                     `json:"selected_food"`
	TypeOverrideByPos map[int]model.FoodType    `json:"type_override_by_pos,omitempty"`
	ExpectedDetails   []model.PetFoodPlanDetail `json:"expected"`
}

func TestCalFeedingAmountPerDay_FromFoodJSON(t *testing.T) {
	allFoods, err := loadFoodJSON(feedingFoodsFile)
	if err != nil {
		t.Fatalf("failed to load %s: %v", feedingFoodsFile, err)
	}
	if len(allFoods) < 4 {
		t.Fatalf("%s must contain at least 4 foods, got %d", feedingFoodsFile, len(allFoods))
	}

	testCases, err := loadFeedingCaseJSON(feedingCasesFile)
	if err != nil {
		t.Fatalf("failed to load %s: %v", feedingCasesFile, err)
	}
	if len(testCases) == 0 {
		t.Fatalf("no feeding test cases found in %s", feedingCasesFile)
	}

	selected := parseCaseSelector(os.Getenv(calFeedingCaseEnv))
	service := &calculationService{}

	ran := 0
	for _, tc := range testCases {
		if len(selected) > 0 {
			if _, ok := selected[tc.Name]; !ok {
				continue
			}
		}

		ran++
		t.Run(tc.Name, func(t *testing.T) {
			foods := pickFoodsForCase(t, allFoods, tc.FoodIndexes, tc.TypeOverrideByPos)
			if len(tc.ExpectedDetails) != len(foods) {
				t.Fatalf("invalid testcase %q: expectedDetails=%d foods=%d", tc.Name, len(tc.ExpectedDetails), len(foods))
			}

			petDetail := &model.PetDetail{Energy: tc.PetEnergy}
			got := service.CalFeedingAmountPerDay(petDetail, foods)
			if len(got) != len(foods) {
				t.Fatalf("expected %d feeding details, got %d", len(foods), len(got))
			}

			t.Logf("case=%q pet_energy=%.10f entries=%d", tc.Name, tc.PetEnergy, len(got))
			for i := range got {
				t.Logf(
					"case=%q idx=%d want={amount:%.10f energy:%.10f protein:%.10f fat:%.10f} got={amount:%.10f energy:%.10f protein:%.10f fat:%.10f}",
					tc.Name,
					i,
					tc.ExpectedDetails[i].Amount,
					tc.ExpectedDetails[i].EnergyIntake,
					tc.ExpectedDetails[i].ProteinIntake,
					tc.ExpectedDetails[i].FatIntake,
					got[i].Amount,
					got[i].EnergyIntake,
					got[i].ProteinIntake,
					got[i].FatIntake,
				)
				assertPetFoodPlanDetailEqual(t, tc.ExpectedDetails[i], got[i], 0.01)
			}
		})
	}

	if len(selected) > 0 && ran == 0 {
		t.Fatalf("no testcase matched %q from %s", os.Getenv(calFeedingCaseEnv), calFeedingCaseEnv)
	}
}

func loadFoodJSON(filename string) ([]model.Food, error) {
	path, err := requirementCSVPath(filename)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var foods []model.Food
	if err := json.Unmarshal(raw, &foods); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", filename, err)
	}

	return foods, nil
}

func loadFeedingCaseJSON(filename string) ([]calFeedingCase, error) {
	path, err := requirementCSVPath(filename)
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cases []calFeedingCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		return nil, fmt.Errorf("unmarshal %s: %w", filename, err)
	}

	return cases, nil
}

func pickFoodsForCase(t *testing.T, allFoods []model.Food, indexes []int, overrideByPos map[int]model.FoodType) []model.Food {
	t.Helper()

	foods := make([]model.Food, 0, len(indexes))
	for pos, idx := range indexes {
		if idx < 0 || idx >= len(allFoods) {
			t.Fatalf("food index out of range: position=%d index=%d max=%d", pos, idx, len(allFoods)-1)
		}
		food := allFoods[idx]
		if overrideType, ok := overrideByPos[pos]; ok {
			food.Type = overrideType
		}
		foods = append(foods, food)
	}
	return foods
}

func assertPetFoodPlanDetailEqual(t *testing.T, expected model.PetFoodPlanDetail, actual *model.PetFoodPlanDetail, tolerance float64) {
	t.Helper()

	if !assert.NotNil(t, actual, "actual PetFoodPlanDetail is nil") {
		return
	}

	assert.Equal(t, normalizePetFoodPlanDetail(expected, tolerance), normalizePetFoodPlanDetail(*actual, tolerance))
}

func normalizePetFoodPlanDetail(detail model.PetFoodPlanDetail, tolerance float64) model.PetFoodPlanDetail {
	detail.Amount = normalizeFloat(detail.Amount, tolerance)
	detail.EnergyIntake = normalizeFloat(detail.EnergyIntake, tolerance)
	detail.ProteinIntake = normalizeFloat(detail.ProteinIntake, tolerance)
	detail.FatIntake = normalizeFloat(detail.FatIntake, tolerance)
	return detail
}

func normalizeFloat(value float64, tolerance float64) float64 {
	if tolerance <= 0 {
		return value
	}
	return math.Round(value/tolerance) * tolerance
}

func parseCaseSelector(raw string) map[string]struct{} {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	selected := map[string]struct{}{}
	for _, name := range strings.Split(raw, ",") {
		cleanName := strings.TrimSpace(name)
		if cleanName == "" {
			continue
		}
		selected[cleanName] = struct{}{}
	}
	return selected
}
