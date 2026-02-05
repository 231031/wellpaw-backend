package service

import (
	"testing"

	"github.com/231031/wellpaw-backend/internal/model"
)

func TestGetNutritientPerDay_FromCSV(t *testing.T) {
	nutritientService := NewNutritientRequirementService()

	rows, err := loadRequirementCSV("requirement_testcase.csv")
	if err != nil {
		t.Fatalf("failed to load test cases: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("no test cases found in CSV")
	}

	for _, row := range rows {
		row := row
		t.Run(requirementRowName(row), func(t *testing.T) {
			t.Logf("case: type=%d sex=%d neutered=%t energy=%.6f expected_protein=%.6f expected_fat=%.6f",
				int(row.petType),
				int(row.sexType),
				row.neutered,
				row.energy,
				row.protein,
				row.fat,
			)
			gotProtein, gotFat := nutritientService.GetNutritientPerDay(model.ADULT, row.petType, row.energy)
			assertInDelta(t, row.protein, gotProtein, 1.0)
			assertInDelta(t, row.fat, gotFat, 1.0)
		})
	}
}
