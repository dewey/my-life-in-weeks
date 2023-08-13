package main

import (
	"embed"
	"flag"
	"fmt"
	"my-life-in-weeks/reverselocation"
	"my-life-in-weeks/service/importing"
	"my-life-in-weeks/store"
	"os"
	"path/filepath"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/peterbourgon/ff/v3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	var (
		environment        = fs.String("environment", "development", "the environment we are running in")
		locationBackend    = fs.String("location-backend", "positionstack", "the location resolver backend we are using")
		positionstackToken = fs.String("positionstack-token", "", "the api token for positionstack")
		databasePath       = fs.String("database-path", "locations.db", "the path to the locations database")
		photoPath          = fs.String("photo-path", "", "the path to the photos that we want to scan")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVars(),
	)
	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	switch strings.ToLower(*environment) {
	case "development":
		l = level.NewFilter(l, level.AllowInfo())
	case "prod":
		l = level.NewFilter(l, level.AllowError())
	}
	l = log.With(l, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	// Connect to sqlite database, create if it doesn't exist and run available migrations if needed
	db, err := sqlx.Open("sqlite3", *databasePath)
	if err != nil {
		level.Error(l).Log("msg", "error opening database", "err", err)
		return
	}
	if err := db.Ping(); err != nil {
		level.Error(l).Log("msg", "error pinging database", "err", err)
		return
	}
	goose.SetBaseFS(embedMigrations)
	f, err := embedMigrations.ReadDir("migrations")
	if err != nil {
		level.Error(l).Log("msg", "error reading migrations directory", "err", err)
		return
	}
	for _, entry := range f {
		level.Info(l).Log("msg", fmt.Sprintf("found migration %s", entry.Name()))
	}

	if err := goose.SetDialect("sqlite"); err != nil {
		level.Error(l).Log("msg", "error setting dialect for database", "err", err)
		return
	}

	if err := goose.Up(db.DB, "migrations"); err != nil {
		level.Error(l).Log("msg", "error running migrations", "err", err)
		return
	}

	var reverseLocationRepository reverselocation.ReverseLocation
	switch *locationBackend {
	case "positionstack":
		reverseLocationRepository = reverselocation.NewRepository(*positionstackToken)
		level.Info(l).Log("msg", "using positionstack as reverse location backend")
	case "in-memory":
		fallthrough
	default:
		reverseLocationRepository = reverselocation.NewInMemoryRepository()
		level.Info(l).Log("msg", "using in-memory mock as reverse location backend")
	}
	storeRepository, err := store.NewRepository(l, db)
	if err != nil {
		level.Error(l).Log("msg", "couldn't initialize store repository", "err", err)
		return
	}
	et, err := exiftool.NewExiftool(exiftool.CoordFormant("%+.10f"))
	if err != nil {
		level.Error(l).Log("msg", "error on initialization of exif tool", "err", err)
		return
	}
	defer et.Close()

	s := importing.NewService(l, storeRepository, reverseLocationRepository, et)
	err = filepath.Walk(*photoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if err := s.Process(path); err != nil {
				level.Debug(l).Log("msg", "error when processing image, skipping file", "err", err)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		level.Error(l).Log("msg", "error processing image", "err", err)
		return
	}
}
