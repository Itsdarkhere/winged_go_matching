package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"wingedapp/pgtester/cmd/wingedapp"
	"wingedapp/pgtester/internal/db/conn"
	dbMigrate "wingedapp/pgtester/internal/db/migrate"
	aiDBMigration "wingedapp/pgtester/internal/wingedapp/aibackend/db/migration"
	wingedappMigration "wingedapp/pgtester/internal/wingedapp/migration"
	supabaseAuthMigration "wingedapp/pgtester/internal/wingedapp/supabase/db/migration"

	"github.com/golang-migrate/migrate/v4"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

const (
	sqlBoiler = "sqlboiler"

	/* internal app */
	appCfgFolder = "./internal/wingedapp/db/pgmodel"
	appCfgPath   = "./cmd/wingedapp/sqlboiler/sqlboiler.toml"

	/* ai app */
	aiCfgFolder = "./internal/wingedapp/aibackend/db/aibckendpgmdl"
	aiCfgPath   = "./cmd/wingedapp/sqlboiler/sqlboiler.ai.toml"

	/* supabase auth */
	supabaseCfgFolder = "./internal/wingedapp/supabase/db/supabasepgmodel"
	supabaseCfgPath   = "./cmd/wingedapp/sqlboiler/sqlboiler.supabase.toml"
)

type Config struct {
	Folder string
	Path   string
	Stmts  embed.FS
}

var (
	cfgs = []Config{
		{
			Folder: appCfgFolder,
			Path:   appCfgPath,
			Stmts:  wingedappMigration.Stmts,
		},
		{
			Folder: aiCfgFolder,
			Path:   aiCfgPath,
			Stmts:  aiDBMigration.Stmts,
		},
		{
			Folder: supabaseCfgFolder,
			Path:   supabaseCfgPath,
			Stmts:  supabaseAuthMigration.Stmts,
		},
	}
)

func main() {
	// db connection
	v := viper.New()
	v.SetEnvPrefix(wingedapp.EnvPrefixDB)
	v.AutomaticEnv()

	port, err := strconv.Atoi(v.GetString(conn.Port))
	if err != nil {
		log.Fatalf("expected no error converting port to int: %v", err)
	}

	connCfg := &conn.Cfg{
		Database:   v.GetString(conn.Database),
		Host:       v.GetString(conn.Host),
		Port:       port,
		User:       v.GetString(conn.User),
		Pass:       v.GetString(conn.Pass),
		Schema:     "public",
		SearchPath: `public, "$user", extensions, backend_app, ai_backend, supabase_auth`,
	}

	db, err := connCfg.PGConn(context.Background(), true)
	if err != nil {
		log.Fatalf("expected no error connecting to MySQL: %v", err)
	}

	// this is an edge case where we:
	//	- recreate a schema (public)
	//	- run migrations on that schema
	//	- generate models from that schema
	//
	// this is so that SQLBoiler will remain schema-agnostic, because if you
	// specify a schema in the config, it will hard-code that schema in the
	// generated models, which is not what we want.
	// TLDR; must be "public"
	for _, cfg := range cfgs {
		if err = recreateSchema(db, "public"); err != nil {
			log.Fatalf("expected no error recreating schema: %v", err)
		}

		if err = migrateDB(cfg.Stmts, connCfg); err != nil {
			log.Fatalf("expected no error migrating DB: %v", err)
		}

		// Then seed those data accordingly,
		generateModelsFromSchema(cfg.Path, cfg.Folder)
	}
}

func migrateDB(embedStmts embed.FS, connCfg *conn.Cfg) error {
	opts := &dbMigrate.Options{
		Stmts:   embedStmts,
		ConnCfg: connCfg,
	}

	if err := dbMigrate.PG(context.Background(), opts); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("no migrations were not run: %w", err)
		}

		return fmt.Errorf("expected no error running migrations for db: '%w'", err)
	}

	return nil
}

// recreateSchema drops and creates the given schema in the provided database.
func recreateSchema(db *sqlx.DB, schema string) error {
	if err := conn.DropPGSchema(context.Background(), db, schema); err != nil {
		return fmt.Errorf("expected no error dropping schema %s: %v", schema, err)
	}

	if err := conn.CreatePGSchema(context.Background(), db, schema); err != nil {
		return fmt.Errorf("expected no error creating schema %s: %v", schema, err)
	}

	return nil
}

func generateModelsFromSchema(cfgFile, cfgFolder string) {
	var (
		err    error
		output []byte
	)

	defer func() {
		if err != nil {
			fmt.Println("====== fail err ======", err)
			fmt.Println("cfgFile:", cfgFile)
			fmt.Println("cfgFolder:", cfgFolder)
		}
	}()

	fmt.Println("Generating models...")

	// ensure config exists
	if _, err = os.Stat(cfgFile); err != nil {
		err = fmt.Errorf("could not find sqlboiler.toml: %v", err)
		return
	}

	// remove existing folder to ensure a clean generation
	if err = os.RemoveAll(cfgFolder); err != nil {
		err = fmt.Errorf("failed to remove folder: %v", err)
		return
	}

	// create file
	fmt.Println("=== going to create file:", cfgFolder)
	if err = os.MkdirAll(cfgFolder, os.ModePerm); err != nil {
		fmt.Println("===== error creating file:", err)
		err = fmt.Errorf("failed to create folder: %v", err)
		return
	}

	// execute generation command
	cmd := exec.Command(sqlBoiler, "psql", "-c", cfgFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error running sqlboiler command: %w", err)
		fmt.Println("===== error running sqlboiler command:", err)
		return
	}

	fmt.Println("==== output:", output)
	fmt.Println("sqlboiler models generated successfully")
}
