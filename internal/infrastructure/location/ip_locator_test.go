package location

import (
	"errors"
	"net"
	"testing"
	"url-shortening-service/internal/domain/mocks"

	"github.com/oschwald/geoip2-golang"
	"github.com/stretchr/testify/assert"
)

func TestLocateIP(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name        string
		ip          string
		expectedRes IPLocation
		expectError bool

		setupMock func(t *testing.T) GeoIP
	}

	testCases := []testCase{
		{
			name: "success - valid IP returns location",
			ip:   "8.8.8.8",
			expectedRes: IPLocation{
				City:    "Mountain View",
				Country: "United States",
			},
			expectError: false,
			setupMock: func(t *testing.T) GeoIP {
				return mocks.NewGeoIpMock(
					func(ip net.IP) (*geoip2.City, error) {
						city := &geoip2.City{}
						city.City.Names = map[string]string{"en": "Mountain View"}
						city.Country.Names = map[string]string{"en": "United States"}
						return city, nil
					},
					func() error {
						return nil
					},
				)
			},
		},
		{
			name: "success - empty location data",
			ip:   "192.168.1.1",
			expectedRes: IPLocation{
				City:    "",
				Country: "",
			},
			expectError: false,
			setupMock: func(t *testing.T) GeoIP {
				return mocks.NewGeoIpMock(
					func(ip net.IP) (*geoip2.City, error) {
						return &geoip2.City{}, nil
					},
					func() error {
						return nil
					},
				)
			},
		},
		{
			name:        "error - geoIP City returns error",
			ip:          "invalid-ip",
			expectedRes: IPLocation{},
			expectError: true,
			setupMock: func(t *testing.T) GeoIP {
				return mocks.NewGeoIpMock(
					func(ip net.IP) (*geoip2.City, error) {
						return nil, errors.New("failed to lookup IP")
					},
					func() error {
						return nil
					},
				)
			},
		},
		{
			name: "success - IPv6 address",
			ip:   "2001:4860:4860::8888",
			expectedRes: IPLocation{
				City:    "San Francisco",
				Country: "United States",
			},
			expectError: false,
			setupMock: func(t *testing.T) GeoIP {
				return mocks.NewGeoIpMock(
					func(ip net.IP) (*geoip2.City, error) {
						city := &geoip2.City{}
						city.City.Names = map[string]string{"en": "San Francisco"}
						city.Country.Names = map[string]string{"en": "United States"}
						return city, nil
					},
					func() error {
						return nil
					},
				)
			},
		},
	}

	for _, tc := range testCases {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockGeoIP := tt.setupMock(t)

			locator := NewGeoIpLocator(mockGeoIP)
			res, err := locator.LocateIP(tt.ip)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
		})
	}
}
