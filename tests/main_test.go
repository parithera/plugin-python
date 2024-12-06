package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	dbhelper "github.com/CodeClarityCE/utility-dbhelper/helper"
	codeclarity "github.com/CodeClarityCE/utility-types/codeclarity_db"
	plugin "github.com/parithera/plugin-star/src"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func TestCreateNPMv1(t *testing.T) {
	os.Setenv("PG_DB_HOST", "127.0.0.1")
	os.Setenv("PG_DB_PORT", "5432")
	os.Setenv("PG_DB_USER", "postgres")
	os.Setenv("PG_DB_PASSWORD", "!ChangeMe!")

	dsn := "postgres://postgres:!ChangeMe!@127.0.0.1:5432/" + dbhelper.Config.Database.Results + "?sslmode=disable"
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn), pgdriver.WithTimeout(50*time.Second)))
	db_codeclarity := bun.NewDB(sqldb, pgdialect.New())
	defer db_codeclarity.Close()

	sourceCodeDir := "/Users/cedric/Documents/workspace/parithera-dev/private/20e14aae-b8ca-4fad-a351-6d747b9ab070/67e09357-aefb-44a2-a978-1c508e16eb23"
	out := plugin.Start(sourceCodeDir, db_codeclarity)

	// Assert the expected values
	assert.NotNil(t, out)
	assert.Equal(t, codeclarity.SUCCESS, out.AnalysisInfo.Status)

	writeJSON(out, sourceCodeDir+"/fastqc.json")
}
