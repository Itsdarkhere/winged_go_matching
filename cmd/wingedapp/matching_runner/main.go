package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"wingedapp/pgtester/internal/wingedapp/apprepo"
	"wingedapp/pgtester/internal/wingedapp/db"
	"wingedapp/pgtester/internal/wingedapp/lib/applog"
	"wingedapp/pgtester/internal/wingedapp/lib/matching"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/extmatcher"
	"wingedapp/pgtester/internal/wingedapp/lib/matching/store"

	"github.com/robfig/cron/v3"
)

func main() {
	// CLI flags for manual triggering
	runDrop := flag.Bool("drop", false, "Run DropOneMatchPerUser once and exit")
	runMatch := flag.Bool("match", false, "Run RunMatchForUnmatchedUsers once and exit")
	populateCSV := flag.String("populate", "", "Populate test users from CSV file path")
	depopulate := flag.Bool("depopulate", false, "Delete all test users (is_test_user=true)")
	flag.Parse()

	cfg := loadConfig()
	logger := applog.NewLogrus("matching-runner")

	// Connect to backend_app DB
	backendDB, err := db.NewTransactor(&db.Config{
		Host:         cfg.DBHost,
		Port:         cfg.DBPort,
		User:         cfg.DBUser,
		Pass:         cfg.DBPass,
		Database:     cfg.DBName,
		Schema:       cfg.DBSchema,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("connect to backend_app db: %v", err)
	}

	// Connect to ai_backend DB (same server, different schema)
	aiBackendDB, err := db.NewTransactor(&db.Config{
		Host:         cfg.DBHost,
		Port:         cfg.DBPort,
		User:         cfg.DBUser,
		Pass:         cfg.DBPass,
		Database:     cfg.DBName,
		Schema:       cfg.DBSchemaAI,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("connect to ai_backend db: %v", err)
	}

	// Connect to supabase auth DB (same server, different schema)
	supabaseDB, err := db.NewTransactor(&db.Config{
		Host:         cfg.DBHost,
		Port:         cfg.DBPort,
		User:         cfg.DBUser,
		Pass:         cfg.DBPass,
		Database:     cfg.DBName,
		Schema:       cfg.DBSchemaSupabase,
		MaxOpenConns: 10,
		MaxIdleConns: 5,
	})
	if err != nil {
		log.Fatalf("connect to supabase auth db: %v", err)
	}

	// Create matching logic with minimal dependencies
	stores := store.NewMatchingStores(logger)
	userDeleter := &apprepo.Store{}
	matchLogic, err := matching.NewLogic(
		stores.MatchConfigStore,
		stores.UserStore,
		stores.UserDatingPrefsStore,
		stores.MatchSetStore,
		stores.MatchResultStore,
		extmatcher.NewQualitativeMatcher(logger),
		stores.ProfileStore,
		stores.SupabaseStore,
		stores.UserMatchActionsStore,
		stores.AudioStore,
		stores.LovestoryStore,
		&noopPublicURLer{},  // not needed for matching logic
		userDeleter,
		stores.DateInstanceStore,
		stores.MatchResultStore,
	)
	if err != nil {
		log.Fatalf("create matching logic: %v", err)
	}

	ctx := context.Background()
	dbExec := backendDB.DB()
	aiExec := aiBackendDB.DB()
	supabaseExec := supabaseDB.DB()

	// Handle --populate: load test users from CSV
	if *populateCSV != "" {
		log.Printf("=== POPULATE MODE START ===")
		log.Printf("loading users from: %s", *populateCSV)

		file, err := os.Open(*populateCSV)
		if err != nil {
			log.Fatalf("open CSV file: %v", err)
		}
		defer file.Close()

		csvResult, err := matching.ParsePopulationCSV(file)
		if err != nil {
			log.Fatalf("parse CSV: %v", err)
		}
		log.Printf("parsed %d valid rows from CSV", csvResult.ValidRows)

		if len(csvResult.ParsingErrors) > 0 {
			log.Printf("parsing errors: %v", csvResult.ParsingErrors)
		}

		execs := &matching.PopulationExecutors{
			BackendApp:   dbExec,
			AIBackend:    aiExec,
			SupabaseAuth: supabaseExec,
		}

		result, err := matchLogic.Populate(ctx, execs, csvResult.Rows, &matching.PopulateOptions{
			IsTestUser: true, // always mark as test users
		})
		if err != nil {
			log.Fatalf("populate users: %v", err)
		}

		log.Printf("populated %d users:", result.TotalProcessed)
		log.Printf("  - backend_app.users: %d", result.BackendAppUsers)
		log.Printf("  - ai_backend.profiles: %d", result.AIBackendProfiles)
		log.Printf("  - dating_preferences: %d", result.DatingPreferences)
		log.Printf("=== POPULATE MODE END ===")
		return
	}

	// Handle --depopulate: delete all test users
	if *depopulate {
		log.Printf("=== DEPOPULATE MODE START ===")
		log.Println("deleting all test users (is_test_user=true)...")

		// Delete from ai_backend.profiles first (references backend_app.users.id)
		res, err := aiExec.ExecContext(ctx, `
			DELETE FROM profiles
			WHERE user_id::text IN (SELECT id::text FROM backend_app.users WHERE is_test_user = true)
		`)
		if err != nil {
			log.Fatalf("delete ai_backend profiles: %v", err)
		}
		profilesDeleted, _ := res.RowsAffected()
		log.Printf("deleted %d ai_backend.profiles", profilesDeleted)

		// Delete dating preferences
		res, err = dbExec.ExecContext(ctx, `
			DELETE FROM user_dating_preferences
			WHERE user_id IN (SELECT id FROM users WHERE is_test_user = true)
		`)
		if err != nil {
			log.Fatalf("delete dating preferences: %v", err)
		}
		prefsDeleted, _ := res.RowsAffected()
		log.Printf("deleted %d dating_preferences", prefsDeleted)

		// Delete match_results referencing test users
		res, err = dbExec.ExecContext(ctx, `
			DELETE FROM match_result
			WHERE initiator_user_ref_id IN (SELECT id FROM users WHERE is_test_user = true)
			   OR receiver_user_ref_id IN (SELECT id FROM users WHERE is_test_user = true)
		`)
		if err != nil {
			log.Fatalf("delete match_results: %v", err)
		}
		matchResultsDeleted, _ := res.RowsAffected()
		log.Printf("deleted %d match_results", matchResultsDeleted)

		// Get supabase_ids before deleting users
		rows, err := dbExec.QueryContext(ctx, `
			SELECT supabase_id FROM users WHERE is_test_user = true AND supabase_id IS NOT NULL
		`)
		if err != nil {
			log.Fatalf("query supabase_ids: %v", err)
		}
		var supabaseIDs []string
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				log.Fatalf("scan supabase_id: %v", err)
			}
			supabaseIDs = append(supabaseIDs, id)
		}
		rows.Close()

		// Delete backend_app users
		res, err = dbExec.ExecContext(ctx, `DELETE FROM users WHERE is_test_user = true`)
		if err != nil {
			log.Fatalf("delete backend_app users: %v", err)
		}
		usersDeleted, _ := res.RowsAffected()
		log.Printf("deleted %d backend_app.users", usersDeleted)

		// Delete from auth.users
		if len(supabaseIDs) > 0 {
			for _, id := range supabaseIDs {
				_, err = supabaseExec.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
				if err != nil {
					log.Printf("warning: failed to delete auth.users %s: %v", id, err)
				}
			}
			log.Printf("deleted %d auth.users", len(supabaseIDs))
		}

		log.Printf("=== DEPOPULATE MODE END ===")
		return
	}

	// Manual trigger mode - run once and exit
	if *runDrop {
		log.Println("manually triggering DropOneMatchPerUser...")
		if err := matchLogic.DropOneMatchPerUser(ctx, dbExec); err != nil {
			log.Fatalf("error dropping matches: %v", err)
		}
		log.Println("DropOneMatchPerUser completed successfully")
		return
	}

	if *runMatch {
		log.Println("=== MATCH MODE START ===")
		log.Println("step 1: calling RunMatchForUnmatchedUsers...")
		matchSet, err := matchLogic.RunMatchForUnmatchedUsers(ctx, dbExec)
		if err != nil {
			log.Fatalf("error running match for unmatched users: %v", err)
		}
		if matchSet == nil {
			log.Println("no match set created (not enough users or no new pairs)")
			log.Println("=== MATCH MODE END ===")
			return
		}
		log.Printf("step 1 done: created match set %s", matchSet.ID)

		log.Println("step 2: calling RunIngestionSet to process matches...")
		if err := matchLogic.RunIngestionSet(ctx, dbExec, aiExec, matchSet.ID); err != nil {
			log.Fatalf("error running ingestion set: %v", err)
		}
		log.Println("step 2 done: RunIngestionSet completed")
		log.Println("=== MATCH MODE END ===")
		return
	}

	// Daemon mode - start cron jobs
	if err := startMatchingCrons(matchLogic, backendDB, aiBackendDB); err != nil {
		log.Fatalf("start matching crons: %v", err)
	}

	log.Println("matching runner started")
	select {} // block forever
}

func startMatchingCrons(matchLogic *matching.Logic, backendDB, aiBackendDB *db.Transactor) error {
	c := cron.New()
	ctx := context.Background()
	dbExec := backendDB.DB()
	aiExec := aiBackendDB.DB()

	// Load match config from DB
	matchCfg, err := matchLogic.MatchConfig(ctx, dbExec, &matching.QueryFilterMatchConfig{})
	if err != nil {
		return fmt.Errorf("load match config: %v", err)
	}

	// Match drops - drop one match per user at configured hours
	for _, dropHour := range matchCfg.DropHours {
		hour := dropHour
		cronExpr := fmt.Sprintf("0 %v * * *", hour)
		_, _ = c.AddFunc(cronExpr, func() {
			log.Printf("running match drops for hour %s", hour)
			if err := matchLogic.DropOneMatchPerUser(ctx, dbExec); err != nil {
				log.Printf("error dropping matches: %v", err)
			}
		})
		log.Printf("scheduled match drops at hour %s", hour)
	}

	// Run matching for unmatched users - daily
	matchHourExpr := fmt.Sprintf("0 %v * * *", matchCfg.MatchExpirationHours)
	_, _ = c.AddFunc(matchHourExpr, func() {
		log.Println("running match for unmatched users")
		matchSet, err := matchLogic.RunMatchForUnmatchedUsers(ctx, dbExec)
		if err != nil {
			log.Printf("error running match for unmatched users: %v", err)
			return
		}
		if matchSet == nil {
			log.Println("no match set created (not enough users or no new pairs)")
			return
		}
		log.Printf("created match set: %s, running matching algorithm...", matchSet.ID)
		if err := matchLogic.RunIngestionSet(ctx, dbExec, aiExec, matchSet.ID); err != nil {
			log.Printf("error running ingestion set: %v", err)
			return
		}
		log.Printf("match set %s processed successfully", matchSet.ID)
	})
	log.Printf("scheduled unmatched users matching at hour %s", matchCfg.MatchExpirationHours)

	c.Start()
	return nil
}

// Config for the matching runner
type Config struct {
	DBHost           string
	DBPort           int
	DBUser           string
	DBPass           string
	DBName           string
	DBSchema         string
	DBSchemaAI       string
	DBSchemaSupabase string
}

func loadConfig() *Config {
	loadDotEnv()

	port, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))

	return &Config{
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           port,
		DBUser:           getEnv("DB_USER", "postgres"),
		DBPass:           getEnv("DB_PASS", ""),
		DBName:           getEnv("DB_NAME", "wingedapp"),
		DBSchema:         getEnv("DB_SCHEMA", "backend_app"),
		DBSchemaAI:       getEnv("DB_SCHEMA_AI", "ai_backend"),
		DBSchemaSupabase: getEnv("DB_SCHEMA_SUPABASE", "auth"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func loadDotEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// noopPublicURLer is a no-op implementation for publicURLer interface
type noopPublicURLer struct{}

func (n *noopPublicURLer) PublicURL(_ context.Context, key string) (string, error) {
	return key, nil // just return the key as-is
}
