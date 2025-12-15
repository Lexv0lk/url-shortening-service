package location

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

const (
	dbPath      = "assets/GeoLite2-City.mmdb"
	defaultLang = "en"
)

type IPLocation struct {
	City    string
	Country string
}

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
