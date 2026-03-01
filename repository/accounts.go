package repository

import (
	"time"

	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/google/uuid"
)

func (r *Repository) GetAllAccounts() ([]models.Account, error) {
	query := "SELECT id, account_name, player, updated_at, created_at FROM accounts WHERE deleted_at IS NULL"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(&acc.ID, &acc.AccountName, &acc.Player, &acc.UpdatedAt, &acc.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r Repository) CreateAccount(accountName string, player string) error {
	query := `
	INSERT INTO accounts (id, account_name, player, created_at, updated_at)
	VALUES(?, ?, ?, ?, ?)
	`

	now := time.Now().UTC().Format(time.RFC3339)

	idString := uuid.New().String()

	_, err := r.db.Exec(query, idString, accountName, player, now, now)
	return err
}

func (r *Repository) GetAccountByID(id string) (models.Account, error) {
	query := `
	SELECT id, account_name, player, updated_at, created_at
	FROM accounts
	WHERE id = ?
	`

	var a models.Account
	err := r.db.QueryRow(query, id).Scan(
		&a.ID,
		&a.AccountName,
		&a.Player,
		&a.UpdatedAt,
		&a.CreatedAt,
	)
	if err != nil {
		return models.Account{}, err
	}
	return a, nil
}

const updateAccount = `
UPDATE accounts
SET account_name = ?, 
		player = ?, 
		updated_at = ?
WHERE id = ?
`

type UpdateAccountParams struct {
	ID          string
	AccountName string
	Player      string
	UpdatedAt   string
}

func (r *Repository) UpdateAccount(arg UpdateAccountParams) error {
	_, err := r.db.Exec(updateAccount,
		arg.AccountName,
		arg.Player,
		arg.UpdatedAt,
		arg.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) SearchAccounts(searchTerm string) ([]models.Account, error) {
	query := `
	SELECT id, account_name, player, updated_at, created_at 
	FROM accounts 
	WHERE deleted_at IS NULL 
	AND (account_name LIKE ? OR player LIKE ?)
	`

	searchPattern := "%" + searchTerm + "%"
	rows, err := r.db.Query(query, searchPattern, searchPattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		err := rows.Scan(&acc.ID, &acc.AccountName, &acc.Player, &acc.UpdatedAt, &acc.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}
