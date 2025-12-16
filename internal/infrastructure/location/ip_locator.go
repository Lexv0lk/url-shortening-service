package location

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

const (
	dbPath      = "assets/GeoLite2-City.mmdb"
	defaultLang = "en"
)

// IPLocation represents geographic location data resolved from an IP address.
type IPLocation struct {
	// City is the city name resolved from the IP address.
	City string
	// Country is the country name resolved from the IP address.
	Country string
}

// LocateIP resolves an IP address to its geographic location using GeoLite2 database.
// It returns the city and country names in English.
//
// Returns an error if:
//   - The GeoLite2 database cannot be opened
//   - The IP address cannot be parsed or located
func LocateIP(ip string) (IPLocation, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return IPLocation{}, err
	}
	defer db.Close()

	parsedIP := net.ParseIP(ip)

	record, err := db.City(parsedIP)
	if err != nil {
		return IPLocation{}, err
	}

	city := record.City.Names[defaultLang]
	country := record.Country.Names[defaultLang]

	return IPLocation{
		City:    city,
		Country: country,
	}, nil
}
