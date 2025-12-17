package location

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

const (
	defaultLang = "en"
)

// GeoIP defines the interface for querying geographic information from an IP address.
type GeoIP interface {
	City(ip net.IP) (*geoip2.City, error)
	Close() error
}

// IPLocator defines the interface for locating geographic information from an IP address.
type IPLocator interface {
	LocateIP(ip string) (IPLocation, error)
}

// GeoIpLocator implements IPLocator using a GeoIP database.
type GeoIpLocator struct {
	geoIP GeoIP
}

// IPLocation represents geographic location data resolved from an IP address.
type IPLocation struct {
	// City is the city name resolved from the IP address.
	City string
	// Country is the country name resolved from the IP address.
	Country string
}

// NewGeoIpLocator creates a new GeoIpLocator with the provided GeoIP interface.
func NewGeoIpLocator(geoIP GeoIP) *GeoIpLocator {
	return &GeoIpLocator{
		geoIP: geoIP,
	}
}

// LocateIP resolves an IP address to its geographic location using GeoLite2 database.
// It returns the city and country names in English.
//
// Returns an error if:
//   - The GeoLite2 database cannot be opened
//   - The IP address cannot be parsed or located
func (l *GeoIpLocator) LocateIP(ip string) (IPLocation, error) {
	parsedIP := net.ParseIP(ip)
	record, err := l.geoIP.City(parsedIP)
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

// OpenGeoIPDatabase opens the GeoLite2 database from the specified file path.
// It returns a GeoIP interface for querying IP location data.
//
// Returns an error if the database cannot be opened.
func OpenGeoIPDatabase(dbPath string) (GeoIP, error) {
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}
