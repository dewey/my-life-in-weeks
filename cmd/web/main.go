package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io/fs"
	_ "io/fs"
	"my-life-in-weeks/store"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	emoji "github.com/jayco/go-emoji-flag"
	"github.com/jmoiron/sqlx"

	"github.com/peterbourgon/ff/v3"
)

//go:embed templates
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	var (
		environment  = flagSet.String("environment", "develop", "the environment we are running in")
		port         = flagSet.String("port", "8080", "the port the web interface is running on")
		databasePath = flagSet.String("database-path", "locations.db", "the path to the locations database")
	)

	ff.Parse(flagSet, os.Args[1:],
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

	if _, err := os.Stat(*databasePath); err != nil {
		level.Error(l).Log("msg", "error opening database. does it exist?", "err", err)
		return
	}
	db, err := sqlx.Open("sqlite3", *databasePath)
	if err != nil {
		level.Error(l).Log("msg", "error opening database", "err", err)
		return
	}
	if err := db.Ping(); err != nil {
		level.Error(l).Log("msg", "error pinging database", "err", err)
		return
	}
	storeRepository, err := store.NewRepository(l, db)
	if err != nil {
		level.Error(l).Log("msg", "couldn't initialize store repository", "err", err)
		return
	}

	r := chi.NewRouter()
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		level.Error(l).Log("msg", "error on getting mounted static sub directory", "err", err)
		return
	}
	// The filesystem is registered on "static/some-file", that's why we have to strip the prefix. It's also important
	// to register the "/static/*" route as otherwise we don't hit it for the individual file names.
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t := template.New("index.html")
		t = template.Must(t.Funcs(template.FuncMap{
			"toEmoji": toEmoji,
		}).ParseFS(templateFS, "templates/index.html"))

		list, err := storeRepository.List()
		if err != nil {
			level.Error(l).Log("msg", "error fetching list data", "err", err)
			http.Error(w, "couldn't fetch list data", 502)
			return
		}
		weeksByYear := make(map[int]map[int]store.Week)
		// Fake row for table heading in the frontend
		weeksByYear[0000] = generateYearWeeks()
		for _, week := range list {
			currentWeek := week
			if _, ok := weeksByYear[currentWeek.Year]; !ok {
				// Every year has 52 weeks
				weeksByYear[currentWeek.Year] = generateYearWeeks()
			} else {
				if _, ok := weeksByYear[currentWeek.Year][week.Number]; ok {
					weeksByYear[currentWeek.Year][week.Number] = store.Week{
						Number:     currentWeek.Number,
						Year:       currentWeek.Year,
						Country:    currentWeek.Country,
						PhotoCount: currentWeek.PhotoCount,
					}
				}
			}
		}

		err = t.Execute(w, weeksByYear)
		if err != nil {
			fmt.Println("err", err)
		}
	})

	//fr := feed.NewRepository(l)
	//cacheRepository, err := cache.NewRepository(l, db)
	//if err != nil {
	//	fmt.Println("err", err)
	//	return
	//}
	//listenerService := hooklistener.NewService(l, fr, notifiers, cacheRepository, *feedURL, *hookToken)
	//
	//r.Mount("/incoming-hooks", hooklistener.NewHandler(*listenerService))
	//
	level.Info(l).Log("msg", fmt.Sprintf("my-life-in-weeks is running on http://localhost:%s", *port), "environment", *environment)

	//Set up webserver
	err = http.ListenAndServe(fmt.Sprintf(":%s", *port), r)
	if err != nil {
		level.Error(l).Log("err", err)
		return
	}
}

func generateYearWeeks() map[int]store.Week {
	initialWeeks := make(map[int]store.Week)
	for i := 0; i < 52; i++ {
		if _, ok := initialWeeks[i+1]; !ok {
			initialWeeks[i+1] = store.Week{Number: i + 1}
		}
	}
	return initialWeeks
}

func toEmoji(country string) string {
	return emoji.GetFlag(country)
}
