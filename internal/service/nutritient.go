package service

import "github.com/231031/pethealth-backend/internal/model"

func getNutritientCat(rangeAge model.AgeType) (float64, float64) {
	protienMap := map[model.AgeType]float64{
		model.JUNIOR: 75,
		model.ADULT:  65,
		model.SENIOR: 65,
	}

	fatMap := map[model.AgeType]float64{
		model.JUNIOR: 22.5,
		model.ADULT:  22.5,
		model.SENIOR: 22.5,
	}

	return protienMap[rangeAge], fatMap[rangeAge]
}

func getNutritientDog(rangeAge model.AgeType) (float64, float64) {
	protienMap := map[model.AgeType]float64{
		model.JUNIOR: 56.3,
		model.ADULT:  45,
		model.SENIOR: 45,
	}

	fatMap := map[model.AgeType]float64{
		model.JUNIOR: 21.3,
		model.ADULT:  13.8,
		model.SENIOR: 13.8,
	}

	return protienMap[rangeAge], fatMap[rangeAge]
}

func GetNutritientPerDay(rangeAge model.AgeType, p model.PetType, mer float64) (float64, float64) {
	var pFactor, fFactor float64
	if p == model.DOG {
		pFactor, fFactor = getNutritientDog(rangeAge)
	} else {
		pFactor, fFactor = getNutritientCat(rangeAge)
	}

	return mer * (pFactor / 1000), mer * (fFactor / 1000)
}
