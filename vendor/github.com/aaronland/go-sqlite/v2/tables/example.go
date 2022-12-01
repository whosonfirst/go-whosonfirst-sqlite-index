package tables

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
)

type ExampleTable struct {
	sqlite.Table
	name string
}

func NewExampleTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewExampleTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewExampleTable(ctx context.Context) (sqlite.Table, error) {

	t := ExampleTable{
		name: "example",
	}

	return &t, nil
}

func (t *ExampleTable) Name() string {
	return t.name
}

func (t *ExampleTable) Schema() string {

	sql := `CREATE TABLE %s (
		id INTEGER NOT NULL,
		body TEXT
	);`

	return fmt.Sprintf(sql, t.Name())
}

func (t *ExampleTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *ExampleTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {

	conn, err := db.Conn(ctx)

	if err != nil {
		return err
	}

	tx, err := conn.Begin()

	if err != nil {
		return err
	}

	b, err := json.Marshal(i)

	if err != nil {
		return err
	}

	id := -1
	body := string(b)

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, body
	) VALUES (
		?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return err
	}

	defer stmt.Close()

	_, err = stmt.Exec(id, body)

	if err != nil {
		return err
	}

	return tx.Commit()
}
