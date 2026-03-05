package repository

import (
	"time"

	"github.com/ByChanderZap/exile-tracker/models"
)

type DashboardStats struct {
	AccountCount   int
	CharacterCount int
	SnapshotCount  int
	FetchQueueSize int
	ActiveFetchers int
}

type RecentCharacter struct {
	models.Character
	AccountName string
}

func (r *Repository) GetDashboardStats() (DashboardStats, error) {
	var stats DashboardStats

	err := r.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL").Scan(&stats.AccountCount)
	if err != nil {
		return stats, err
	}

	err = r.db.QueryRow("SELECT COUNT(*) FROM characters WHERE deleted_at IS NULL").Scan(&stats.CharacterCount)
	if err != nil {
		return stats, err
	}

	err = r.db.QueryRow("SELECT COUNT(*) FROM pobsnapshots WHERE deleted_at IS NULL").Scan(&stats.SnapshotCount)
	if err != nil {
		return stats, err
	}

	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM characters_to_fetch ctf
		JOIN characters c ON c.id = ctf.character_id
		WHERE ctf.should_skip = 0 AND c.disabled = 0 AND c.deleted_at IS NULL
	`).Scan(&stats.FetchQueueSize)
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func (r *Repository) GetRecentlyUpdatedCharacters(limit int) ([]RecentCharacter, error) {
	query := `
	SELECT c.id, c.account_id, c.character_name, c.died, c.disabled, c.current_league,
	       c.created_at, c.updated_at, a.account_name
	FROM characters c
	INNER JOIN accounts a ON a.id = c.account_id
	WHERE c.deleted_at IS NULL
	ORDER BY c.updated_at DESC
	LIMIT ?
	`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []RecentCharacter
	for rows.Next() {
		var rc RecentCharacter
		var updatedAt string
		var createdAt string
		err := rows.Scan(
			&rc.ID,
			&rc.AccountId,
			&rc.CharacterName,
			&rc.Died,
			&rc.Disabled,
			&rc.CurrentLeague,
			&createdAt,
			&updatedAt,
			&rc.AccountName,
		)
		if err != nil {
			return nil, err
		}
		rc.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		rc.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		chars = append(chars, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return chars, nil
}
