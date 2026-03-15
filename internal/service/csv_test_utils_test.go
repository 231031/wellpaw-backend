package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/231031/wellpaw-backend/internal/model"
)

type requirementCSVRow struct {
	petType  model.PetType
	sexType  model.SexType
	neutered bool
	weight   float64
	bcsScore int
	bcs      model.BcsType
	energy   float64
	protein  float64
	fat      float64
}

func loadRequirementCSV(filename string) ([]requirementCSVRow, error) {
	csvPath, err := requirementCSVPath(filename)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	idx, err := columnIndex(header, []string{
		"type",
		"sex",
		"reproduction",
		"weight",
		"bcs",
		"energy",
		"protein",
		"fat",
	})
	if err != nil {
		return nil, err
	}

	var rows []requirementCSVRow
	for rowNum := 2; ; rowNum++ {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read row %d: %w", rowNum, err)
		}
		if isEmptyRow(record) {
			continue
		}

		petTypeRaw, err := parseInt(valueAt(record, idx["type"]))
		if err != nil {
			return nil, fmt.Errorf("row %d type: %w", rowNum, err)
		}
		petType := model.PetType(petTypeRaw)

		sexRaw, err := parseInt(valueAt(record, idx["sex"]))
		if err != nil {
			return nil, fmt.Errorf("row %d sex: %w", rowNum, err)
		}
		sexType := model.SexType(sexRaw)

		reproductionRaw, err := parseInt(valueAt(record, idx["reproduction"]))
		if err != nil {
			return nil, fmt.Errorf("row %d reproduction: %w", rowNum, err)
		}
		neutered := reproductionRaw == int(model.NEUTERED)

		weight, err := parseFloat(valueAt(record, idx["weight"]))
		if err != nil {
			return nil, fmt.Errorf("row %d weight: %w", rowNum, err)
		}

		bcsScore, err := parseInt(valueAt(record, idx["bcs"]))
		if err != nil {
			return nil, fmt.Errorf("row %d bcs: %w", rowNum, err)
		}
		bcs := mapBcsScore(bcsScore)

		energy, err := parseFloat(valueAt(record, idx["energy"]))
		if err != nil {
			return nil, fmt.Errorf("row %d energy: %w", rowNum, err)
		}
		protein, err := parseFloat(valueAt(record, idx["protein"]))
		if err != nil {
			return nil, fmt.Errorf("row %d protein: %w", rowNum, err)
		}
		fat, err := parseFloat(valueAt(record, idx["fat"]))
		if err != nil {
			return nil, fmt.Errorf("row %d fat: %w", rowNum, err)
		}

		rows = append(rows, requirementCSVRow{
			petType:  petType,
			sexType:  sexType,
			neutered: neutered,
			weight:   weight,
			bcsScore: bcsScore,
			bcs:      bcs,
			energy:   energy,
			protein:  protein,
			fat:      fat,
		})
	}

	return rows, nil
}

func requirementCSVPath(filename string) (string, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to locate test file path")
	}
	serviceDir := filepath.Dir(thisFile)
	rootDir := filepath.Dir(filepath.Dir(serviceDir))
	return filepath.Join(rootDir, filename), nil
}

func requirementRowName(row requirementCSVRow) string {
	return fmt.Sprintf("type%d_sex%d_neutered%d_w%.2f_bcs%d",
		int(row.petType),
		int(row.sexType),
		boolToInt(row.neutered),
		row.weight,
		row.bcsScore,
	)
}

func columnIndex(header []string, required []string) (map[string]int, error) {
	index := make(map[string]int, len(required))
	headerMap := make(map[string]int, len(header))
	for i, col := range header {
		headerMap[strings.TrimSpace(col)] = i
	}
	for _, col := range required {
		i, ok := headerMap[col]
		if !ok {
			return nil, fmt.Errorf("missing required column %q", col)
		}
		index[col] = i
	}
	return index, nil
}

func valueAt(record []string, idx int) string {
	if idx < 0 || idx >= len(record) {
		return ""
	}
	return strings.TrimSpace(record[idx])
}

func isEmptyRow(record []string) bool {
	for _, v := range record {
		if strings.TrimSpace(v) != "" {
			return false
		}
	}
	return true
}

func mapBcsScore(score int) model.BcsType {
	switch score {
	case 1, 2:
		return model.VERYTHIN
	case 3, 4:
		return model.THIN
	case 5:
		return model.IDEAL
	case 6, 7:
		return model.OVERWEIGHT
	case 8, 9:
		return model.OBESITY
	default:
		return model.IDEAL
	}
}

func parseFloat(value string) (float64, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	return strconv.ParseFloat(value, 64)
}

func parseInt(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}

func assertInDelta(t *testing.T, expected, actual, tolerance float64) {
	t.Helper()
	if math.Abs(expected-actual) > tolerance {
		t.Fatalf("expected %.4f, got %.4f (tolerance %.2f)", expected, actual, tolerance)
	}
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
