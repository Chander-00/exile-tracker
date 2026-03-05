package repository

import (
	"time"

	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/google/uuid"
)

const searchCharactersInAccount = `
	SELECT id, character_name
	FROM characters
	WHERE deleted_at IS NULL
	AND account_id = ?
	AND character_name LIKE ?
`

type SearchCharactersInAccountParams struct {
	AccountId string
	Query     string
}

func (r *Repository) SearchCharactersInAccount(params SearchCharactersInAccountParams) ([]models.Character, error) {
	searchPattern := "%" + params.Query + "%"
	rows, err := r.db.Query(searchCharactersInAccount, params.AccountId, searchPattern)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []models.Character

	for rows.Next() {
		var c models.Character
		err := rows.Scan(&c.ID, &c.CharacterName)
		if err != nil {
			return nil, err
		}
		characters = append(characters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *Repository) GetCharactersByAccountId(accountId string) ([]models.Character, error) {
	query := `
	SELECT id, account_id, character_name, died, disabled, current_league, created_at, updated_at
	FROM characters
	WHERE account_id = ? AND deleted_at IS NULL
	`
	rows, err := r.db.Query(query, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []models.Character
	for rows.Next() {
		var char models.Character
		err := rows.Scan(
			&char.ID,
			&char.AccountId,
			&char.CharacterName,
			&char.Died,
			&char.Disabled,
			&char.CurrentLeague,
			&char.CreatedAt,
			&char.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		characters = append(characters, char)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *Repository) CreateCharacter(accountId string, characterName string, currentLeague string) (string, error) {
	query := `
		INSERT INTO characters(id, account_id, character_name, current_league, created_at, updated_at)
		VALUES(?,?,?,?,?,?)
	`

	now := time.Now().UTC().Format(time.RFC3339)

	idString := uuid.New().String()
	_, err := r.db.Exec(query, idString, accountId, characterName, currentLeague, now, now)

	return idString, err
}

func (r *Repository) UpdateDiedStatus(characterId string, died bool) error {
	query := `
		UPDATE characters SET died = ?, updated_at = ? 
		WHERE id = ?
	`

	now := time.Now().UTC().Format(time.RFC3339)

	_, err := r.db.Exec(query, died, now, characterId)

	return err
}

func (r *Repository) KillCharacter(characterId string) error {
	query := `
		UPDATE characters SET died = ?, updated_at = ? 
		WHERE id = ?
	`
	_, err := r.db.Exec(query, true, time.Now().UTC().Format(time.RFC3339), characterId)
	return err
}

func (r *Repository) ToggleDisabled(characterId string) error {
	query := `
		UPDATE characters SET disabled = NOT disabled, updated_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, time.Now().UTC().Format(time.RFC3339), characterId)
	return err
}

func (r *Repository) SoftDeleteCharacter(id string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	// Soft delete all snapshots for this character
	_, err := r.db.Exec(`
		UPDATE pobsnapshots SET deleted_at = ?
		WHERE character_id = ? AND deleted_at IS NULL
	`, now, id)
	if err != nil {
		return err
	}

	// Remove from fetch queue
	_, err = r.db.Exec(`DELETE FROM characters_to_fetch WHERE character_id = ?`, id)
	if err != nil {
		return err
	}

	// Soft delete the character
	_, err = r.db.Exec(`
		UPDATE characters SET deleted_at = ?
		WHERE id = ?
	`, now, id)
	return err
}

func (r *Repository) GetDeletedCharactersByAccountId(accountId string) ([]models.Character, error) {
	query := `
	SELECT id, account_id, character_name, died, disabled, current_league, created_at, updated_at
	FROM characters
	WHERE account_id = ? AND deleted_at IS NOT NULL
	`
	rows, err := r.db.Query(query, accountId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []models.Character
	for rows.Next() {
		var char models.Character
		err := rows.Scan(
			&char.ID,
			&char.AccountId,
			&char.CharacterName,
			&char.Died,
			&char.Disabled,
			&char.CurrentLeague,
			&char.CreatedAt,
			&char.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		characters = append(characters, char)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return characters, nil
}

func (r *Repository) RestoreCharacter(id string) error {
	// Restore the character
	_, err := r.db.Exec(`UPDATE characters SET deleted_at = NULL WHERE id = ?`, id)
	if err != nil {
		return err
	}

	// Restore its snapshots
	_, err = r.db.Exec(`UPDATE pobsnapshots SET deleted_at = NULL WHERE character_id = ?`, id)
	if err != nil {
		return err
	}

	// Re-add to fetch queue if not already there
	_, err = r.db.Exec(`
		INSERT INTO characters_to_fetch (id, character_id)
		SELECT hex(randomblob(16)), ?
		WHERE ? NOT IN (SELECT character_id FROM characters_to_fetch)
	`, id, id)
	return err
}

func (r *Repository) GetCharactersToFetch() ([]models.CharactersToFetch, error) {
	query := `
		SELECT ctf.id, ctf.character_id, ctf.last_fetch, ctf.should_skip
		FROM characters_to_fetch ctf
		JOIN characters c ON c.id = ctf.character_id
		WHERE c.disabled = 0 AND c.deleted_at IS NULL
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cToFetch []models.CharactersToFetch
	for rows.Next() {
		var char models.CharactersToFetch
		err := rows.Scan(
			&char.Id,
			&char.CharacterId,
			&char.LastFetch,
			&char.ShouldSkip,
		)
		if err != nil {
			return nil, err
		}
		cToFetch = append(cToFetch, char)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cToFetch, nil
}

const addCharacterToFetch = `
INSERT INTO characters_to_fetch(id, character_id)
		VALUES(?,?)
`

type AddCharactersToFetchParams struct {
	CharacterId string
}

func (r *Repository) AddCharacterToFetch(params AddCharactersToFetchParams) error {
	id := uuid.New().String()
	_, err := r.db.Exec(addCharacterToFetch, id, params.CharacterId)
	return err
}

func (r *Repository) SetShouldSkip(shouldSkip bool, id string) error {
	query := `
		UPDATE characters_to_fetch
		SET should_skip = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(query, shouldSkip, time.Now().UTC().Format(time.RFC3339), id)
	return err
}

func (r *Repository) GetCharacterByID(id string) (models.Character, error) {
	query := `
    SELECT id, account_id, character_name, died, disabled, current_league, created_at, updated_at
    FROM characters
    WHERE id = ?
    `
	var c models.Character
	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.AccountId,
		&c.CharacterName,
		&c.Died,
		&c.Disabled,
		&c.CurrentLeague,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return models.Character{}, err
	}
	return c, nil
}

func (r *Repository) GetAllCharacters() ([]models.Character, error) {
	query := `
	SELECT id, account_id, character_name, died, disabled, current_league, created_at, updated_at
	FROM characters
	WHERE deleted_at IS NULL
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var characters []models.Character
	for rows.Next() {
		var char models.Character
		err := rows.Scan(
			&char.ID,
			&char.AccountId,
			&char.CharacterName,
			&char.Died,
			&char.Disabled,
			&char.CurrentLeague,
			&char.CreatedAt,
			&char.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		characters = append(characters, char)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return characters, nil
}

const updateCharacter = `
UPDATE characters
SET character_name = ?, 
	died = ?, 
	current_league = ?,
	updated_at = ?
WHERE id = ?
`

type UpdateCharacterParams struct {
	ID            string
	CharacterName string
	Died          bool
	CurrentLeague string
	UpdatedAt     string
}

func (r *Repository) UpdateCharacter(arg UpdateCharacterParams) error {
	_, err := r.db.Exec(updateCharacter,
		arg.CharacterName,
		arg.Died,
		arg.CurrentLeague,
		arg.UpdatedAt,
		arg.ID,
	)
	if err != nil {
		return err
	}
	return nil
}
