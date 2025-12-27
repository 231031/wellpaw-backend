package service

import (
	"math"

	"github.com/231031/wellpaw-backend/internal/model"
)

type RerOperation func(bw float64) float64

func getReproductionFactor(neutered bool, p model.PetType) float64 {
	catMap := map[bool]float64{
		false: 1.4,
		true:  1.2,
	}

	dogMap := map[bool]float64{
		false: 1.8,
		true:  1.6,
	}

	if p == model.DOG {
		return dogMap[neutered]
	}
	return catMap[neutered]
}

func getActivityBcsFactor(al model.ActivityLevel, bcs model.BcsType) (float64, float64) {
	alMap := map[model.ActivityLevel]float64{
		model.INACTIVE:   1,
		model.SOMEACTIVE: 1.2,
		model.ACTIVE:     1.4,
		model.VERYACTIVE: 1.6,
	}

	bcsMap := map[model.BcsType]float64{
		model.VERYTHIN:   1.2,
		model.THIN:       1.2,
		model.IDEAL:      1,
		model.OVERWEIGHT: 0.8,
		model.OBESITY:    0.8,
	}

	return alMap[al], bcsMap[bcs]
}

func RERJuniorCat(bw float64) float64 {
	bwTerm := math.Pow(bw, 0.67)
	return 70 * bwTerm * 2
}

func RERAdultCat(bw float64) float64 {
	bwTerm := math.Pow(bw, 0.67)
	return 70 * bwTerm
}

func RERJuniorDog(bw float64) float64 {
	bwTerm := math.Pow(bw, 0.75)
	return 70 * bwTerm * 2
}

func RERAdultDog(bw float64) float64 {
	bwTerm := math.Pow(bw, 0.75)
	return 70 * bwTerm
}

func getEnergyFormulaDog(key model.AgeType) RerOperation {
	energyFormula := map[model.AgeType]RerOperation{
		model.JUNIOR: RERJuniorCat,
		model.ADULT:  RERAdultCat,
		model.SENIOR: RERAdultCat,
	}
	return energyFormula[key]
}

func getEnergyFormulaCat(key model.AgeType) RerOperation {
	energyFormula := map[model.AgeType]RerOperation{
		model.JUNIOR: RERJuniorDog,
		model.ADULT:  RERAdultDog,
		model.SENIOR: RERAdultDog,
	}
	return energyFormula[key]
}

func GetMerEnergy(bw float64, rangeAge model.AgeType, al model.ActivityLevel, bcs model.BcsType, neutered bool, pet model.PetType) float64 {
	var rerFormula RerOperation
	if pet == model.DOG {
		rerFormula = getEnergyFormulaDog(rangeAge)
	} else {
		rerFormula = getEnergyFormulaCat(rangeAge)
	}

	if rangeAge == model.JUNIOR {
		return rerFormula(bw)
	}
	alFactor, bcsFator := getActivityBcsFactor(al, bcs)
	neuteredFactor := getReproductionFactor(neutered, pet)
	return rerFormula(bw) * alFactor * bcsFator * neuteredFactor
}
