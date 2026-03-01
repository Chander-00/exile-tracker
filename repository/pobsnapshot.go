package repository

import (
	"time"

	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/google/uuid"
)

const createPobSnapshot = `
INSERT INTO pobsnapshots (id, character_id, export_string, created_at, updated_at)
	VALUES(?, ?, ?, ?, ?)
`

type CreatePoBSnapshotParams struct {
	CharacterId  string
	ExportString string
}

func (r *Repository) CreatePOBSnapshot(params CreatePoBSnapshotParams) error {
	now := time.Now().UTC().Format(time.RFC3339)
	idString := uuid.New().String()

	_, err := r.db.Exec(createPobSnapshot,
		idString,
		params.CharacterId,
		params.ExportString,
		now,
		now,
	)
	return err
}

const getSnapshotsByCharacterWithExtras = `
	SELECT p.id, p.export_string, c.character_name, a.account_name, p.created_at  
	FROM pobsnapshots p
	INNER JOIN characters c on c.id = p.character_id
	INNER JOIN accounts a on a.id = c.account_id
	WHERE p.character_id = ?
	ORDER BY p.created_at DESC
`

type GetSnapshotsByCharacterWithExtras struct {
	CharacterId string
}

func (r *Repository) GetSnapshotsByCharacterWithExtras(params GetSnapshotsByCharacterWithExtras) ([]models.SnapshotWithExtras, error) {
	rows, err := r.db.Query(getSnapshotsByCharacterWithExtras, params.CharacterId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var swe []models.SnapshotWithExtras
	for rows.Next() {

		var s models.SnapshotWithExtras
		err := rows.Scan(
			&s.SnapshotData.ID,
			&s.SnapshotData.ExportString,
			&s.CharacterName,
			&s.AccountName,
			&s.SnapshotData.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		swe = append(swe, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return swe, nil
}

func (r *Repository) GetSnapshotsByCharacter(characterId string) ([]models.POBSnapshot, error) {
	query := `
	SELECT id, character_id, export_string, created_at, updated_at, deleted_at
	FROM pobsnapshots
	WHERE character_id = ?
	ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, characterId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.POBSnapshot
	for rows.Next() {
		var s models.POBSnapshot
		err := rows.Scan(
			&s.ID,
			&s.CharacterId,
			&s.ExportString,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (r *Repository) GetLatestSnapshotByCharacter(characterId string) (models.POBSnapshot, error) {
	query := `
	SELECT id, character_id, export_string, created_at, updated_at, deleted_at
	FROM pobsnapshots 
	WHERE character_id = ?
	ORDER BY created_at DESC
	LIMIT 1
	`
	var s models.POBSnapshot
	err := r.db.QueryRow(query, characterId).Scan(
		&s.ID,
		&s.CharacterId,
		&s.ExportString,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.DeletedAt,
	)
	if err != nil {
		return models.POBSnapshot{}, err
	}
	return s, nil
}

func (r *Repository) GetSnapshotByID(id string) (models.POBSnapshot, error) {
	query := `
	SELECT id, character_id, export_string, created_at, updated_at, deleted_at
	FROM pobsnapshots
	WHERE id = ?
	`
	var s models.POBSnapshot
	err := r.db.QueryRow(query, id).Scan(
		&s.ID,
		&s.CharacterId,
		&s.ExportString,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.DeletedAt,
	)
	if err != nil {
		return models.POBSnapshot{}, err
	}
	return s, nil
}
