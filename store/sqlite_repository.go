package store

import (
	"encoding/json"
	"time"

	"github.com/go-kit/log"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	l  log.Logger
	db *sqlx.DB
}

// NewRepository initializes a new store repository backed by sqlite
func NewRepository(l log.Logger, db *sqlx.DB) (*repository, error) {
	return &repository{
		l:  l,
		db: db,
	}, nil
}

// Get returns a cache entry for a given key
func (s *repository) Get(path string) (*Entry, bool, error) {
	var entry Entry
	err := s.db.Get(&entry, "select * from photos where path=$1", path)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &entry, true, nil
}

// Add adds a new importing entry
func (s *repository) Add(path string, coordinates Coordinates, meta MetaInformation, geo GeoInformation) error {
	b, err := json.Marshal(geo)
	if err != nil {
		return err
	}
	_, err = s.db.NamedExec("insert into photos (path, longitude, latitude, country, city, original_created_at, data, created_at) values (:path, :longitude, :latitude, :country, :city, :original_created_at, :data, :created_at) on conflict(path) do nothing",
		map[string]interface{}{
			"path":                path,
			"longitude":           coordinates.Longitude,
			"latitude":            coordinates.Latitude,
			"country":             geo.Country,
			"city":                geo.City,
			"original_created_at": meta.OriginalCreatedAt,
			"data":                string(b),
			"created_at":          time.Now(),
		})
	return err
}

// Exists checks if a metadata entry for a specific path exists
func (s *repository) Exists(path string) (bool, error) {
	var count int
	if err := s.db.Get(&count, "select count(*) from photos where path=$1", path); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *repository) List() ([]Week, error) {
	var weeks []Week
	if err := s.db.Select(&weeks, "select country, strftime('%Y',original_created_at) as year, strftime('%W',original_created_at) as number, count(*) as photo_count from photos where country != \"\" and strftime('%Y',original_created_at) > \"1900\" group by 1, 2, 3 having count(*) > 5 order by 1, 2, 3 desc;"); err != nil {
		return nil, err
	}
	return weeks, nil
}
