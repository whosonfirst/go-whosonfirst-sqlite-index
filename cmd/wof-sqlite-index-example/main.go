package main

import (
	"context"
	"flag"
	"fmt"
	wof_index "github.com/whosonfirst/go-whosonfirst-index"
	_ "github.com/whosonfirst/go-whosonfirst-index-csv"
	_ "github.com/whosonfirst/go-whosonfirst-index-sqlite"
	"github.com/whosonfirst/go-whosonfirst-log"
	"github.com/whosonfirst/go-whosonfirst-sqlite"
	"github.com/whosonfirst/go-whosonfirst-sqlite-index"
	"github.com/whosonfirst/go-whosonfirst-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite/tables"
	"github.com/whosonfirst/go-reader"
	_ "github.com/whosonfirst/go-reader-http"	
	"io"
	"os"
	"strings"
	"time"
)

type Example struct {
	Time int64 `json:"time"`
}

func main() {

	valid_modes := strings.Join(wof_index.Modes(), ",")
	desc_modes := fmt.Sprintf("The mode to use importing data. Valid modes are: %s.", valid_modes)

	mode := flag.String("mode", "files", desc_modes)

	dsn := flag.String("dsn", ":memory:", "")
	driver := flag.String("driver", "sqlite3", "")

	live_hard := flag.Bool("live-hard-die-fast", true, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")

	index_belongsto := flag.Bool("index-belongs-to", false, "Index wof:belongsto records for each feature")
	belongsto_reader_uri := flag.String("belongs-to-uri", "", "A valid whosonfirst/go-reader Reader URI for reading wof:belongsto records")

	flag.Parse()

	ctx := context.Background()
	
	logger := log.SimpleWOFLogger()

	stdout := io.Writer(os.Stdout)
	logger.AddLogger(stdout, "status")

	db, err := database.NewDBWithDriver(*driver, *dsn)

	if err != nil {
		logger.Fatal("unable to create database (%s) because %s", *dsn, err)
	}

	defer db.Close()

	if *live_hard {

		err = db.LiveHardDieFast()

		if err != nil {
			logger.Fatal("Unable to live hard and die fast so just dying fast instead, because %s", err)
		}
	}

	to_index := make([]sqlite.Table, 0)

	ex, err := tables.NewExampleTableWithDatabase(db)

	if err != nil {
		logger.Fatal("failed to create 'example' table because '%s'", err)
	}

	to_index = append(to_index, ex)

	cb := func(ctx context.Context, fh io.Reader, args ...interface{}) (interface{}, error) {

		now := time.Now()

		e := Example{
			Time: now.Unix(),
		}

		return e, nil
	}

	idx_opts := &index.SQLiteIndexerOptions{
		DB: db,
		Tables: to_index,
		Callback: cb,
	}

	if *index_belongsto {

		r, err := reader.NewReader(ctx, *belongsto_reader_uri)

		if err != nil {
			logger.Fatal("Failed to create Reader (%s), %v", *belongsto_reader_uri, err)
		}

		idx_opts.IndexBelongsTo = true
		idx_opts.BelongsToReader = r
	}
	
	idx, err := index.NewSQLiteIndexer(idx_opts)

	if err != nil {
		logger.Fatal("failed to create sqlite indexer because %s", err)
	}

	idx.Timings = *timings
	idx.Logger = logger

	err = idx.IndexPaths(*mode, flag.Args())

	if err != nil {
		logger.Fatal("Failed to index paths in %s mode because: %s", *mode, err)
	}

	os.Exit(0)
}
