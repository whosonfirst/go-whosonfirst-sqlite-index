package database

import (
	"fmt"
	"net/url"
)

func DSNFromURI(uri string) (string, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return "", fmt.Errorf("Failed to parse URI, %w", err)
	}

	host := u.Host
	path := u.Path
	q := u.RawQuery

	/*

		if !strings.HasPrefix(dsn, "file:") {

			// because this and this:

			if dsn == ":memory:" {

				// https://github.com/mattn/go-sqlite3#faq
				// https://github.com/mattn/go-sqlite3/issues/204

				dsn = "file::memory:?mode=memory&cache=shared"

			} else if strings.HasPrefix(dsn, "vfs:") {

				// see also: https://github.com/aaronland/go-sqlite-vfs
				// pass

			} else {

				// https://github.com/mattn/go-sqlite3/issues/39
				dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc", dsn)

			}
		}
	*/

	var dsn string

	if host == "mem" {
		dsn = "file::memory:?mode=memory&cache=shared"
	} else {
		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc", path)
	}

	if q != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, q)
	}

	return dsn, nil
}
