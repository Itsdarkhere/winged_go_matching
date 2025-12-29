package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"strconv"
	dbConn "wingedapp/pgtester/internal/db/conn"
	dbMigrate "wingedapp/pgtester/internal/db/migrate"
	aiDBMigration "wingedapp/pgtester/internal/wingedapp/aibackend/db/migration"
	wingedappMigration "wingedapp/pgtester/internal/wingedapp/migration"
	supabaseAuthMigration "wingedapp/pgtester/internal/wingedapp/supabase/db/migration"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/viper"
)

const (
	wingedConnPrefix = "WINGED_PG"
)

var (
	stmts = wingedappMigration.Stmts
)

type migrateEmbed struct {
	schemaName string
	stmts      embed.FS
}

func main() {
	v := viper.New()
	v.SetEnvPrefix(wingedConnPrefix)
	v.AutomaticEnv()

	port, err := strconv.Atoi(v.GetString(dbConn.Port))
	if err != nil {
		log.Fatalf("expected no error converting port to int: %v", err)
	}

	connCfg := &dbConn.Cfg{
		Database:   v.GetString(dbConn.Database),
		Host:       v.GetString(dbConn.Host),
		Port:       port,
		User:       v.GetString(dbConn.User),
		Pass:       v.GetString(dbConn.Pass),
		Schema:     "public",
		SearchPath: `"$user", extension, public, backend_app, ai_backend, supabase_auth`,
	}

	var mEbeds = []migrateEmbed{
		{
			schemaName: "public",
			stmts:      wingedappMigration.Stmts,
		},
		{
			schemaName: "backend_app",
			stmts:      wingedappMigration.Stmts,
		},
		{
			schemaName: "ai_backend",
			stmts:      aiDBMigration.Stmts,
		},
		{
			schemaName: "supabase_auth",
			stmts:      supabaseAuthMigration.Stmts,
		},
	}

	for _, me := range mEbeds {
		connCfg.Schema = me.schemaName // manual db override

		opts := &dbMigrate.Options{
			Stmts:   me.stmts,
			ConnCfg: connCfg,
		}

		if err = dbMigrate.PG(context.Background(), opts); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("no migrations were not run")
				return
			}

			log.Fatalf("expected no error running migrations for db: '%s': %v", me.schemaName, err)
		}

		fmt.Println("==== migrated db:", me.schemaName, "====")
	}

	fmt.Println("==== All DBs Migrated Successfully ====")
}
