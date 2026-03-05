package models

import "time"

type SnapshotWithExtras struct {
	SnapshotData  POBSnapshot
	CharacterName string
	AccountName   string
}

type Account struct {
	ID          string  `json:"id"`
	AccountName string  `json:"account_name"`
	Player      *string `json:"player"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type Character struct {
	ID            string  `json:"id"`
	AccountId     string  `json:"account_id"`
	CharacterName string  `json:"CharacterName"`
	Died          bool    `json:"died"`
	Disabled      bool    `json:"disabled"`
	CurrentLeague *string `json:"current_league"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type POBSnapshot struct {
	ID           string `json:"id"`
	CharacterId  string `json:"character_id"`
	ExportString string `json:"export_string"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

type CharactersToFetch struct {
	Id          string     `json:"id"`
	CharacterId string     `json:"character_id"`
	LastFetch   *time.Time `json:"last_fetch"`
	ShouldSkip  bool       `json:"should_skip"`
}
