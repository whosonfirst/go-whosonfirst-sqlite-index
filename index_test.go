package index

import (
	"context"
	"fmt"
	_ "github.com/aaronland/go-sqlite-modernc"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/aaronland/go-sqlite/v2/tables"
	"io"
	"os"
	"testing"
	"time"
)

type Example struct {
	Time int64 `json:"time"`
}

func TestIndexing(t *testing.T) {

	ctx := context.Background()

	db_uri := "modernc://mem"

	db, err := sqlite.NewDatabase(ctx, db_uri)

	if err != nil {
		t.Fatalf("Unable to create database (%s) because %v", db_uri, err)
	}

	defer db.Close(ctx)

	ex_t, err := tables.NewExampleTableWithDatabase(ctx, db)

	if err != nil {
		t.Fatalf("Failed to create example table, %v", err)
	}

	to_index := []sqlite.Table{ex_t}

	record_func := func(ctx context.Context, path string, r io.ReadSeeker, args ...interface{}) (interface{}, error) {

		now := time.Now()

		e := Example{
			Time: now.Unix(),
		}

		fmt.Printf("Index %s\n", path)
		return e, nil
	}

	idx_opts := &SQLiteIndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
	}

	idx, err := NewSQLiteIndexer(idx_opts)

	if err != nil {
		t.Fatalf("Failed to create sqlite indexer because %v", err)
	}

	cwd, err := os.Getwd()

	if err != nil {
		t.Fatalf("Failed to get current working directory, %v", err)
	}

	err = idx.IndexURIs(ctx, "directory://", cwd)

	if err != nil {
		t.Fatalf("Failed to index paths, %v", err)
	}

}
