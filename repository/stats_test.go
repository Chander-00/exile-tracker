package repository

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *Repository {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	schema := `
	CREATE TABLE accounts (
		id TEXT PRIMARY KEY,
		account_name TEXT NOT NULL,
		player TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP
	);
	CREATE TABLE characters (
		id TEXT PRIMARY KEY,
		account_id TEXT,
		character_name TEXT NOT NULL,
		died BOOLEAN NOT NULL DEFAULT false,
		disabled BOOLEAN NOT NULL DEFAULT false,
		current_league TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP,
		FOREIGN KEY(account_id) REFERENCES accounts(id)
	);
	CREATE TABLE pobsnapshots (
		id TEXT PRIMARY KEY,
		character_id TEXT NOT NULL,
		export_string TEXT NOT NULL,
		pob_code TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL,
		deleted_at TIMESTAMP,
		FOREIGN KEY(character_id) REFERENCES characters(id)
	);
	CREATE TABLE characters_to_fetch (
		id TEXT PRIMARY KEY,
		character_id TEXT NOT NULL,
		last_fetch TEXT,
		should_skip BOOLEAN NOT NULL DEFAULT false,
		updated_at TEXT,
		FOREIGN KEY(character_id) REFERENCES characters(id)
	);
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return NewRepository(db)
}

func seedTestData(t *testing.T, repo *Repository) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339)
	older := time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339)

	stmts := []string{
		`INSERT INTO accounts (id, account_name, player, created_at, updated_at) VALUES ('a1', 'TestAccount1', 'Player1', '` + older + `', '` + now + `')`,
		`INSERT INTO accounts (id, account_name, player, created_at, updated_at) VALUES ('a2', 'TestAccount2', 'Player2', '` + older + `', '` + now + `')`,
		`INSERT INTO characters (id, account_id, character_name, died, current_league, created_at, updated_at) VALUES ('c1', 'a1', 'Char1', 0, 'League1', '` + older + `', '` + now + `')`,
		`INSERT INTO characters (id, account_id, character_name, died, current_league, created_at, updated_at) VALUES ('c2', 'a1', 'Char2', 1, 'League1', '` + older + `', '` + older + `')`,
		`INSERT INTO characters (id, account_id, character_name, died, current_league, created_at, updated_at) VALUES ('c3', 'a2', 'Char3', 0, 'League2', '` + older + `', '` + older + `')`,
		`INSERT INTO pobsnapshots (id, character_id, export_string, created_at, updated_at) VALUES ('s1', 'c1', 'export1', '` + now + `', '` + now + `')`,
		`INSERT INTO pobsnapshots (id, character_id, export_string, created_at, updated_at) VALUES ('s2', 'c1', 'export2', '` + older + `', '` + older + `')`,
		`INSERT INTO characters_to_fetch (id, character_id, should_skip) VALUES ('f1', 'c1', 0)`,
		`INSERT INTO characters_to_fetch (id, character_id, should_skip) VALUES ('f2', 'c2', 1)`,
	}

	for _, stmt := range stmts {
		if _, err := repo.db.Exec(stmt); err != nil {
			t.Fatalf("failed to seed data: %v\nstmt: %s", err, stmt)
		}
	}
}

func TestGetDashboardStats(t *testing.T) {
	repo := setupTestDB(t)
	seedTestData(t, repo)

	stats, err := repo.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats failed: %v", err)
	}

	if stats.AccountCount != 2 {
		t.Errorf("expected 2 accounts, got %d", stats.AccountCount)
	}
	if stats.CharacterCount != 3 {
		t.Errorf("expected 3 characters, got %d", stats.CharacterCount)
	}
	if stats.SnapshotCount != 2 {
		t.Errorf("expected 2 snapshots, got %d", stats.SnapshotCount)
	}
	if stats.FetchQueueSize != 1 {
		t.Errorf("expected 1 in fetch queue (non-skipped), got %d", stats.FetchQueueSize)
	}
}

func TestGetDashboardStatsEmpty(t *testing.T) {
	repo := setupTestDB(t)

	stats, err := repo.GetDashboardStats()
	if err != nil {
		t.Fatalf("GetDashboardStats failed: %v", err)
	}

	if stats.AccountCount != 0 || stats.CharacterCount != 0 || stats.SnapshotCount != 0 || stats.FetchQueueSize != 0 {
		t.Errorf("expected all zeros for empty db, got %+v", stats)
	}
}

func TestGetRecentlyUpdatedCharacters(t *testing.T) {
	repo := setupTestDB(t)
	seedTestData(t, repo)

	chars, err := repo.GetRecentlyUpdatedCharacters(2)
	if err != nil {
		t.Fatalf("GetRecentlyUpdatedCharacters failed: %v", err)
	}

	if len(chars) != 2 {
		t.Fatalf("expected 2 characters, got %d", len(chars))
	}

	// First should be most recently updated (c1 has the latest updated_at)
	if chars[0].CharacterName != "Char1" {
		t.Errorf("expected first char to be Char1, got %s", chars[0].CharacterName)
	}
	if chars[0].AccountName != "TestAccount1" {
		t.Errorf("expected account name TestAccount1, got %s", chars[0].AccountName)
	}
}

func TestGetRecentlyUpdatedCharactersEmpty(t *testing.T) {
	repo := setupTestDB(t)

	chars, err := repo.GetRecentlyUpdatedCharacters(5)
	if err != nil {
		t.Fatalf("GetRecentlyUpdatedCharacters failed: %v", err)
	}

	if len(chars) != 0 {
		t.Errorf("expected 0 characters, got %d", len(chars))
	}
}
