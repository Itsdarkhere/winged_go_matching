package debugger

import (
	"context"
	"fmt"
	"log"
	"wingedapp/pgtester/internal/util/strutil"

	"github.com/jmoiron/sqlx"
)

func PGTables(ctx context.Context, db *sqlx.DB) {
	fmt.Println("====== pg tables:")

	const q = `
SELECT tablename
FROM pg_catalog.pg_tables
WHERE schemaname NOT IN ('pg_catalog','information_schema')
ORDER BY tablename;
`
	var names []string
	if err := db.SelectContext(ctx, &names, q); err != nil {
		log.Fatalf("pg_tables query failed: %v", err)
	}

	fmt.Println("====== names:", strutil.GetAsJson(names))
	fmt.Println("====== end pg tables:")
}
