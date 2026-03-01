package models

type CreateCharacterInput struct {
	AccountId     string `json:"account_id" validate:"required"`
	CharacterName string `json:"CharacterName" validate:"required"`
	Died          bool   `json:"died"`
	CurrentLeague string `json:"current_league" validate:"required"`
}

type UpdateCharacterInput struct {
	CharacterName string `json:"character_name" validate:"required"`
	CurrentLeague string `json:"current_league" validate:"required"`
}

type AddCharacterToFetchInput struct {
	CharacterId string `json:"character_id" validate:"required"`
}
