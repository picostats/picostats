package main

import (
	"encoding/json"
	"log"
	"os"
)

type TimeZone struct {
	Value  string   `json:"value"`
	Abbr   string   `json:"abbr"`
	Offset float64  `json:"offset"`
	IsDst  bool     `json:"isdst"`
	Text   string   `json:"text"`
	UTC    []string `json:"utc"`
}

type TimeZonesManager struct {
	TimeZones []*TimeZone `json:"timezones"`
}

func (tzm *TimeZonesManager) load(configFile string) error {
	file, err := os.Open(configFile)

	if err != nil {
		log.Printf("[zones.go] Error while opening config file: %v", err)
		return err
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&tzm)

	if err != nil {
		log.Printf("[zones.go] Error while decoding JSON: %v", err)
		return err
	}

	return nil
}

func (tzm *TimeZonesManager) getOffset(zone string) float64 {
	return 0.0
}

func initZones() *TimeZonesManager {
	tzm := &TimeZonesManager{}
	tzm.load("timezones.json")
	return tzm
}
