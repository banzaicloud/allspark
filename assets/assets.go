package assets

import _ "embed"

var (
	// AirportCodes contains airport codes and location data
	//go:embed airport-codes.csv
	AirportCodes []byte
)
