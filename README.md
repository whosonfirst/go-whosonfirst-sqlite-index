# go-whosonfirst-sqlite-index

Go package for indexing SQLite databases (using go-whosonfirst-index).

## Important

This is a "version 2" release and is backwards incompatible with previous versions of this package. If you need to use the older version specify it as follows in your `go.mod` file:

```
require (
	github.com/whosonfirst/go-whosonfirst-sqlite-index v0.2.0
)
```

Documentation for this package is incomplete and will be updated shortly.

## Dependencies and relationships

These are documented in the [Dependencies and relationships section](https://github.com/whosonfirst/go-whosonfirst-sqlite#dependencies-and-relationships) of the `go-whosonfirst-sqlite` package.

## Tools

```
$> make cli
go build -mod vendor -o bin/example cmd/example/main.go
```

### example

```
$> ./bin/example -h
Usage of ./bin/example:
  -driver string
    	 (default "sqlite3")
  -dsn string
    	 (default ":memory:")
  -emitter-uri string
    	The mode to use importing data. Valid modes are: directory://,featurecollection://,file://,filelist://,geojsonl://,repo://. (default "repo://")
  -live-hard-die-fast
    	Enable various performance-related pragmas at the expense of possible (unlikely) database corruption (default true)
  -post-index
    	Enable post indexing callback function
  -timings
    	Display timings during and after indexing
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-sqlite
