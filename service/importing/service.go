package importing

import (
	"encoding/json"
	"errors"
	"fmt"
	"personal-location-history/reverselocation"
	"personal-location-history/store"
	"time"

	"github.com/barasher/go-exiftool"

	"github.com/go-kit/log"

	"github.com/go-kit/log/level"
)

type Service struct {
	l   log.Logger
	sr  store.Store
	rlr reverselocation.ReverseLocation
	et  *exiftool.Exiftool
}

func NewService(l log.Logger, sr store.Store, rlr reverselocation.ReverseLocation, et *exiftool.Exiftool) *Service {
	return &Service{
		l:   l,
		sr:  sr,
		rlr: rlr,
		et:  et,
	}
}

func (s *Service) Process(path string) error {
	ln := log.With(s.l, "path", path)
	level.Info(ln).Log("msg", "processing image")
	exists, err := s.sr.Exists(path)
	if err != nil {
		level.Error(ln).Log("msg", "error checking existence of entry", "err", err)
		return err
	}
	if exists {
		return nil
	}

	fileInfos := s.et.ExtractMetadata(path)
	for _, fileInfo := range fileInfos {
		if fileInfo.Err != nil {
			level.Error(ln).Log("msg", "error getting file info", "err", err)
			return err
		}

		// Skip entry if we don't have geo information in the image
		var hasGPSInformation bool
		if _, ok := fileInfo.Fields["GPSLatitude"]; ok {
			if _, ok := fileInfo.Fields["GPSLongitude"]; ok {
				hasGPSInformation = true
			}
		}
		if !hasGPSInformation {
			return errors.New("input file doesn't have required latitude and longitude information")
		}

		coordinates := store.Coordinates{
			Longitude: fmt.Sprintf("%v", fileInfo.Fields["GPSLongitude"]),
			Latitude:  fmt.Sprintf("%v", fileInfo.Fields["GPSLatitude"]),
		}
		r, err := s.rlr.Address(coordinates.String())
		if err != nil {
			level.Error(ln).Log("msg", "error getting reverse geocoding response", "coordinates", coordinates, "err", err)
			return err
		}

		if r.CountryCode == "" {
			level.Error(ln).Log("msg", "error getting country code", "coordinates", coordinates, "err", err)
			return err
		}

		b, err := json.Marshal(r)
		if err != nil {
			level.Error(ln).Log("msg", "error marshalling reverse geocoding response", "err", err)
			return err
		}

		geo := store.GeoInformation{
			Country:     r.CountryCode,
			City:        r.Region,
			FetchedAt:   time.Now(),
			RawResponse: json.RawMessage(b),
		}

		// Both values are valid, some older file only have one of them
		var originalCreatedAt time.Time
		if _, ok := fileInfo.Fields["DateTimeOriginal"]; ok {
			t, err := time.Parse("2006:01:02 15:04:05", fmt.Sprintf("%s", fileInfo.Fields["DateTimeOriginal"]))
			if err == nil {
				originalCreatedAt = t
			}
		}
		if _, ok := fileInfo.Fields["CreateDate"]; ok {
			t, err := time.Parse("2006:01:02 15:04:05", fmt.Sprintf("%s", fileInfo.Fields["CreateDate"]))
			if err == nil && originalCreatedAt.IsZero() {
				originalCreatedAt = t
			}
		}

		if originalCreatedAt.Before(time.Date(1900, 0, 0, 0, 0, 0, 0, time.UTC)) {
			level.Error(ln).Log("msg", "error when validating original_created_at", "err", errors.New("original_created_at can't be older than 1900-01-01"))
			return err
		}

		if err := s.sr.Add(path, coordinates, store.MetaInformation{OriginalCreatedAt: originalCreatedAt}, geo); err != nil {
			level.Error(ln).Log("msg", "error adding to store", "err", err)
			return err
		}
	}
	return nil
}
