package index

import (
	"context"
	"github.com/aaronland/go-sqlite"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/emitter"
	"github.com/whosonfirst/go-whosonfirst-iterate/v2/iterator"
	"github.com/whosonfirst/go-whosonfirst-log"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

type SQLiteIndexerPostIndexFunc func(context.Context, sqlite.Database, []sqlite.Table, interface{}) error

type SQLiteIndexerLoadRecordFunc func(context.Context, string, io.ReadSeeker, ...interface{}) (interface{}, error)

type SQLiteIndexer struct {
	callback      emitter.EmitterCallbackFunc
	table_timings map[string]time.Duration
	mu            *sync.RWMutex
	Timings       bool
	Logger        *log.WOFLogger
}

type SQLiteIndexerOptions struct {
	DB             sqlite.Database
	Tables         []sqlite.Table
	LoadRecordFunc SQLiteIndexerLoadRecordFunc
	PostIndexFunc  SQLiteIndexerPostIndexFunc
}

func NewSQLiteIndexer(opts *SQLiteIndexerOptions) (*SQLiteIndexer, error) {

	db := opts.DB
	tables := opts.Tables
	record_func := opts.LoadRecordFunc

	table_timings := make(map[string]time.Duration)
	mu := new(sync.RWMutex)

	logger := log.SimpleWOFLogger()

	emitter_cb := func(ctx context.Context, path string, fh io.ReadSeeker, args ...interface{}) error {

		record, err := record_func(ctx, path, fh, args...)

		if err != nil {
			logger.Warning("failed to load record (%s) because %s", path, err)
			return err
		}

		if record == nil {
			return nil
		}

		db.Lock()

		defer db.Unlock()

		for _, t := range tables {

			t1 := time.Now()

			err = t.IndexRecord(ctx, db, record)

			if err != nil {
				logger.Warning("failed to index feature (%s) in '%s' table because %s", path, t.Name(), err)
				return err
			}

			t2 := time.Since(t1)

			n := t.Name()

			mu.Lock()

			_, ok := table_timings[n]

			if ok {
				table_timings[n] += t2
			} else {
				table_timings[n] = t2
			}

			mu.Unlock()
		}

		if opts.PostIndexFunc != nil {

			err := opts.PostIndexFunc(ctx, db, tables, record)

			if err != nil {
				return err
			}
		}

		return nil
	}

	i := SQLiteIndexer{
		callback:      emitter_cb,
		table_timings: table_timings,
		mu:            mu,
		Timings:       false,
		Logger:        logger,
	}

	return &i, nil
}

func (idx *SQLiteIndexer) IndexPaths(ctx context.Context, emitter_uri string, uris []string) error {

	iter, err := iterator.NewIterator(ctx, emitter_uri, idx.callback)

	if err != nil {
		return err
	}

	done_ch := make(chan bool)
	t1 := time.Now()

	// ideally this could be a proper stand-along package method but then
	// we have to set up a whole bunch of scaffolding just to pass 'indexer'
	// around so... we're not doing that (20180205/thisisaaronland)

	show_timings := func() {

		t2 := time.Since(t1)

		i := atomic.LoadInt64(&iter.Seen)

		idx.mu.RLock()
		defer idx.mu.RUnlock()

		for t, d := range idx.table_timings {
			idx.Logger.Status("time to index %s (%d) : %v", t, i, d)
		}

		idx.Logger.Status("time to index all (%d) : %v", i, t2)
	}

	if idx.Timings {

		go func() {

			for {

				select {
				case <-done_ch:
					return
				case <-time.After(1 * time.Minute):
					show_timings()
				}
			}
		}()

		defer func() {
			done_ch <- true
		}()
	}

	err = iter.IterateURIs(ctx, uris...)

	if err != nil {
		return err
	}

	return nil
}