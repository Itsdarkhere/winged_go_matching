package testhelper

import (
	"context"
	"os"
	"testing"
	matching "wingedapp/pgtester/internal/wingedapp/lib/matching"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

/*
	This helper ingests a population of users from a CSV file into the database.
	It uses the canonical matching.Logic.Populate method for all database insertions.
*/

// DBProvider provides database access. Implemented by testsuite.Helper.
type DBProvider interface {
	BackendAppDb() *sqlx.DB
	AiBackendDb() *sqlx.DB
	SupabaseAuthDb() *sqlx.DB
}

// PopulationIngestor ingests population data from CSV files using matching.Logic.Populate.
type PopulationIngestor struct {
	t        *testing.T
	db       DBProvider
	matchLib *matching.Logic
}

// NewPopulationIngestor creates an ingestor for testing.
func NewPopulationIngestor(t *testing.T, db DBProvider, matchLib *matching.Logic) *PopulationIngestor {
	return &PopulationIngestor{t: t, db: db, matchLib: matchLib}
}

// IngestFromCSVFile ingests users from a CSV file using matching.Logic.Populate.
// Returns the parsed rows and populate result for assertions.
func (i *PopulationIngestor) IngestFromCSVFile(file string) (*matching.PopulationCSVResult, *matching.PopulateResult, error) {
	f, err := os.Open(file)
	if err != nil {
		if i.t != nil {
			require.NoError(i.t, err, "opening population file")
		}
		return nil, nil, err
	}
	defer f.Close()

	// Parse CSV using canonical parser
	parseResult, err := matching.ParsePopulationCSV(f)
	if err != nil {
		if i.t != nil {
			require.NoError(i.t, err, "parsing population CSV")
		}
		return nil, nil, err
	}

	execs := &matching.PopulationExecutors{
		BackendApp:   i.db.BackendAppDb(),
		AIBackend:    i.db.AiBackendDb(),
		SupabaseAuth: i.db.SupabaseAuthDb(),
	}

	// Populate using canonical Logic.Populate (nil options = default behavior)
	populateResult, err := i.matchLib.Populate(context.Background(), execs, parseResult.Rows, nil)
	if err != nil {
		if i.t != nil {
			require.NoError(i.t, err, "populating users")
		}
		return parseResult, nil, err
	}

	return parseResult, populateResult, nil
}

// IngestFromCSVFileSimple is a simplified version that just returns row count for backward compatibility.
// Use IngestFromCSVFile for full access to results.
func (i *PopulationIngestor) IngestFromCSVFileSimple(file string) (int, error) {
	parseResult, _, err := i.IngestFromCSVFile(file)
	if err != nil {
		return 0, err
	}
	return parseResult.ValidRows, nil
}

// ProdDBProvider wraps database connections for production use.
type ProdDBProvider struct {
	backendApp   *sqlx.DB
	aiBackend    *sqlx.DB
	supabaseAuth *sqlx.DB
}

func (p *ProdDBProvider) BackendAppDb() *sqlx.DB   { return p.backendApp }
func (p *ProdDBProvider) AiBackendDb() *sqlx.DB    { return p.aiBackend }
func (p *ProdDBProvider) SupabaseAuthDb() *sqlx.DB { return p.supabaseAuth }

// Seeder is a production-ready wrapper for PopulationIngestor.
type Seeder struct {
	ingestor *PopulationIngestor
}

// SeedFromCSVFile seeds users from a CSV file and returns the count of seeded users.
func (s *Seeder) SeedFromCSVFile(ctx context.Context, file string) (int, error) {
	return s.ingestor.IngestFromCSVFileSimple(file)
}
