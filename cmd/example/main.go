package main

import (
	"context"
	"flag"
	"fmt"
	_ "github.com/aaronland/go-sqlite-modernc"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/aaronland/go-sqlite/v2/tables"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-sqlite-index/v4"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

type Example struct {
	Time int64 `json:"time"`
}

func main() {

	valid_modes := strings.Join(emitter.Schemes(), ",")
	desc_modes := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-iterate/v2 URI. Valid schemes are: %s.", valid_modes)

	emitter_uri := flag.String("emitter-uri", "repo://", desc_modes)

	db_uri := flag.String("database-uri", "modernc://mem", "")

	live_hard := flag.Bool("live-hard-die-fast", true, "Enable various performance-related pragmas at the expense of possible (unlikely) database corruption")
	timings := flag.Bool("timings", false, "Display timings during and after indexing")

	post_index := flag.Bool("post-index", false, "Enable post indexing callback function")

	flag.Parse()

	ctx := context.Background()

	db, err := sqlite.NewDatabase(ctx, *db_uri)

	if err != nil {
		log.Fatalf("unable to create database (%s) because %s", *db_uri, err)
	}

	defer db.Close(ctx)

	if *live_hard {

		err = sqlite.LiveHardDieFast(ctx, db)

		if err != nil {
			log.Fatalf("Unable to live hard and die fast so just dying fast instead, because %s", err)
		}
	}

	to_index := make([]sqlite.Table, 0)

	ex, err := tables.NewExampleTableWithDatabase(ctx, db)

	if err != nil {
		log.Fatalf("failed to create 'example' table because '%s'", err)
	}

	to_index = append(to_index, ex)

	record_func := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) (interface{}, error) {

		now := time.Now()

		e := Example{
			Time: now.Unix(),
		}

		return e, nil
	}

	idx_opts := &index.SQLiteIndexerOptions{
		DB:             db,
		Tables:         to_index,
		LoadRecordFunc: record_func,
	}

	if *post_index {

		post_func := func(ctx context.Context, db sqlite.Database, tables []sqlite.Table, record interface{}) error {
			log.Printf("Post index func w/ %v", record)
			return nil
		}

		idx_opts.PostIndexFunc = post_func
	}

	idx, err := index.NewSQLiteIndexer(idx_opts)

	if err != nil {
		log.Fatalf("failed to create sqlite indexer because %s", err)
	}

	idx.Timings = *timings

	err = idx.IndexPaths(ctx, *emitter_uri, flag.Args())

	if err != nil {
		log.Fatalf("Failed to index paths in %s mode because: %s", *emitter_uri, err)
	}

	os.Exit(0)
}
