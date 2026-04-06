package model

type PetUpdateType int

const (
	WEIGHT_UPDATE PetUpdateType = iota
	BCS_UPDATE
	AL_UPDATE
)
