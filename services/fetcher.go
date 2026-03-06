package services

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ByChanderZap/exile-tracker/buildsSitesClient"
	"github.com/ByChanderZap/exile-tracker/config"
	"github.com/ByChanderZap/exile-tracker/models"
	"github.com/ByChanderZap/exile-tracker/poeclient"
	"github.com/ByChanderZap/exile-tracker/repository"
	"github.com/ByChanderZap/exile-tracker/utils"
	"github.com/rs/zerolog"
)

type FetcherService struct {
	repo      *repository.Repository
	poeClient *poeclient.POEClient
	log       zerolog.Logger
	ticker    *time.Ticker
	done      chan struct{}
}

func NewFetcherService(repo *repository.Repository, poeClient *poeclient.POEClient, interval time.Duration) *FetcherService {
	log := utils.ChildLogger("fetcher")
	log.Info().Msgf("Updates will be generated every %s", interval)
	return &FetcherService{
		repo:      repo,
		poeClient: poeClient,
		log:       log,
		ticker:    time.NewTicker(interval),
		done:      make(chan struct{}, 1),
	}
}

func (fs *FetcherService) Start(ctx context.Context) {
	fs.log.Info().Msg("Starting fetcher service")

	// Run once first
	go fs.fetchAllData(ctx)

	go func() {
		for {
			select {
			case <-fs.done:
				fs.log.Info().Msg("Fetcher service stopped")
				return
			case <-fs.ticker.C:
				fs.fetchAllData(ctx)
			case <-ctx.Done():
				fs.log.Info().Msg("Context cancelled, stopping fetcher service")
				return
			}
		}
	}()
}

func (fs *FetcherService) Stop() {
	fs.ticker.Stop()
	close(fs.done)
}

func (fs *FetcherService) fetchAllData(ctx context.Context) {
	fs.log.Info().Msg("Starting fetch cycle")

	charactersToFetch, err := fs.repo.GetCharactersToFetch()
	if err != nil {
		fs.log.Error().Err(err).Msg("Failed to get characters to fetch from database")
		return
	}

	fs.log.Info().Int("characters_to_fetch", len(charactersToFetch)).Msg("Found characters to process")

	for _, ctf := range charactersToFetch {
		select {
		case <-ctx.Done():
			fs.log.Info().Msg("Context cancelled during fetch cycle, stopping")
			return
		default:
		}
		fs.FetchCharacterData(ctx, ctf)
		time.Sleep(2 * time.Second)
	}
	fs.log.Info().Msg("Data fetch cycle completed")
}

func (fs *FetcherService) FetchCharacterData(ctx context.Context, ctf models.CharactersToFetch) {
	c, err := fs.repo.GetCharacterByID(ctf.CharacterId)
	if err != nil {
		fs.log.Error().Err(err).Str("character_id", ctf.CharacterId).Msg("Failed to fetch")
		return
	}

	acc, err := fs.repo.GetAccountByID(c.AccountId)
	if err != nil {
		fs.log.Error().Err(err).Str("account_id", c.AccountId).Msg("Failed to fetch")
		return
	}

	log := fs.log.With().
		Str("account", acc.AccountName).
		Str("character", c.CharacterName).Logger()

	if c.Died {
		log.Warn().Msg("Character is dead, skipping fetch")
		fs.repo.SetShouldSkip(true, ctf.Id)
		return
	}

	items, err := fs.poeClient.GetItemsJson(acc.AccountName, c.CharacterName, "pc")
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch items")
		return
	}

	passives, err := fs.poeClient.GetPassiveSkillsJson(acc.AccountName, c.CharacterName, "pc")
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch passive skills")
		return
	}

	var itemsResponse models.ItemsResponse
	if err := json.Unmarshal(items, &itemsResponse); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall items")
		return
	}

	var passivesResponse models.PassiveSkillsResponse
	if err := json.Unmarshal(passives, &passivesResponse); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshall passive skills")
		return
	}

	err = fs.CreateSnapshot(ctx, c.ID, itemsResponse, passivesResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create snapshot")
		return
	}
}

func (fs *FetcherService) CreateSnapshot(ctx context.Context, characterId string, items models.ItemsResponse, passives models.PassiveSkillsResponse) error {
	dirPath, err := os.MkdirTemp("", "exile-"+characterId+"-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(dirPath)

	itemsPath := filepath.Join(dirPath, "items.json")
	file, err := os.OpenFile(itemsPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error trying to open items file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(items)
	if err != nil {
		return fmt.Errorf("failed to encode json items: %w", err)
	}

	passivesPath := filepath.Join(dirPath, "passives.json")
	passivesFile, err := os.OpenFile(passivesPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error trying to open passives file: %w", err)
	}
	defer passivesFile.Close()

	encoder2 := json.NewEncoder(passivesFile)
	err = encoder2.Encode(passives)
	if err != nil {
		return fmt.Errorf("failed to encode json passives: %w", err)
	}

	buildCode, err := fs.generatePoBBin(ctx, itemsPath, passivesPath)
	if err != nil {
		return fmt.Errorf("failed to execute PoB: %w", err)
	}

	dbSnapshot, err := fs.repo.GetLatestSnapshotByCharacter(characterId)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("failed to get latest snapshot: %w", err)
		}
		fs.log.Warn().Msg("No previous snapshots found.")
	}

	if buildCode.RawCode == dbSnapshot.PobCode {
		fs.log.Info().Msg("No changes detected between latest and current snapshot")
		return nil
	}

	err = fs.repo.CreatePOBSnapshot(repository.CreatePoBSnapshotParams{
		CharacterId:  characterId,
		ExportString: buildCode.URL,
		PobCode:      buildCode.RawCode,
	})
	if err != nil {
		return fmt.Errorf("failed to store snapshot: %w", err)
	}

	return nil
}

type BuildResult struct {
	RawCode string
	URL     string
}

func (fs *FetcherService) generatePoBBin(ctx context.Context, itemsPath string, passivesPath string) (BuildResult, error) {
	fs.log.Info().Msg("Executing Path of Building in headless mode")
	pobRoot := config.Envs.POBRoot

	srcDir := filepath.Join(pobRoot, "src")
	runtimeLua := filepath.Join(pobRoot, "runtime", "lua")
	runtime := filepath.Join(pobRoot, "runtime")

	luajitPath := config.Envs.LuajitPath
	if luajitPath == "" {
		var err error
		luajitPath, err = exec.LookPath("luajit")
		if err != nil {
			return BuildResult{}, fmt.Errorf("luajit not found in PATH (set LUAJIT_PATH to specify it manually): %w", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, luajitPath, "HeadlessWrapper.lua", itemsPath, passivesPath)
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(),
		"LUA_PATH="+runtimeLua+"/?.lua;"+runtimeLua+"/?/init.lua;;",
		"LUA_CPATH="+runtime+"/?.so;"+runtime+"/?.dll;;",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fs.log.Error().Str("stderr", stderr.String()).Msg("PoB execution timed out after 30s")
			return BuildResult{}, fmt.Errorf("PoB execution timed out after 30s: %w", err)
		}
		return BuildResult{}, fmt.Errorf("PoB execution failed: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return BuildResult{}, fmt.Errorf("PoB produced empty output")
	}

	// Take the last non-empty line as the build code
	lines := strings.Split(output, "\n")
	rawCode := strings.TrimSpace(lines[len(lines)-1])

	uploadedBuild, err := buildsSitesClient.UploadBuildWithFallback(rawCode)
	if err != nil {
		return BuildResult{}, fmt.Errorf("failed when uploading build: %w", err)
	}
	fs.log.Debug().Msg(uploadedBuild)
	return BuildResult{RawCode: rawCode, URL: uploadedBuild}, nil
}
