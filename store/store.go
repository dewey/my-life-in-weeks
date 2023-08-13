package store

import (
	"encoding/json"
	"fmt"
	"time"
)

type Store interface {
	Add(path string, coordinates Coordinates, meta MetaInformation, geo GeoInformation) error
	Get(path string) (*Entry, bool, error)
	List() ([]Week, error)
	Exists(path string) (bool, error)
}

// Entry is a importing entry in the store
type Entry struct {
	Path string
	Coordinates
	GeoInformation
}

// Week represents a week
type Week struct {
	Number     int    `json:"number,omitempty" db:"number"`
	Year       int    `json:"year,omitempty" db:"year"`
	Country    string `json:"country,omitempty" db:"country"`
	PhotoCount int    `json:"photo_count,omitempty" db:"photo_count"`
}

// Coordinates holds the decimal representation of the coordinates
type Coordinates struct {
	Longitude string
	Latitude  string
}

func (c *Coordinates) String() string {
	return fmt.Sprintf("%s, %s", c.Latitude, c.Longitude)
}

// GeoInformation is the additional geo metadata we fetched
type GeoInformation struct {
	Country     string
	City        string
	FetchedAt   time.Time
	RawResponse json.RawMessage
}

type MetaInformation struct {
	OriginalCreatedAt time.Time
}
