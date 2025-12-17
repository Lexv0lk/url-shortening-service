package mocks

import (
	"net"

	"github.com/oschwald/geoip2-golang"
)

type GeoIpMock struct {
	cityFn  func(ip net.IP) (*geoip2.City, error)
	closeFn func() error
}

func (g *GeoIpMock) City(ip net.IP) (*geoip2.City, error) {
	return g.cityFn(ip)
}

func (g *GeoIpMock) Close() error {
	return g.closeFn()
}

func NewGeoIpMock(cityFn func(ip net.IP) (*geoip2.City, error), closeFn func() error) *GeoIpMock {
	return &GeoIpMock{
		cityFn:  cityFn,
		closeFn: closeFn,
	}
}
