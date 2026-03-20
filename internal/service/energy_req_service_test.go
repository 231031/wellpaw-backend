package service

import (
	"testing"
	"time"
)

func TestGetMerEnergy_FromCSV(t *testing.T) {
	service := NewEnergyRequirementService()
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
			t.Logf("case: al=%d type=%d sex=%d neutered=%t weight=%.2f bcs=%d expected_energy=%.6f",
				int(row.activityLevel),
				int(row.petType),
				int(row.sexType),
				row.neutered,
				row.weight,
				row.bcsScore,
				row.energy,
			)
			got := service.GetMerEnergy(
				row.weight,
				row.ageRange,
				row.activityLevel,
				row.bcs,
				false,
				time.Time{},
				false,
				row.neutered,
				row.petType,
			)
			assertInDelta(t, row.energy, got, 1.0)
		})
	}
}
