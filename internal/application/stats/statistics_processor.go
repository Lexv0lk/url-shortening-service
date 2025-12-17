package stats

import (
	"context"
	"encoding/json"
	"url-shortening-service/internal/domain"
	"url-shortening-service/internal/infrastructure/location"

	"github.com/mileusna/useragent"
)

const (
	unknowsStr = "Unknown"
	mobileStr  = "Mobile"
	desktopStr = "Desktop"
	tabletStr  = "Tablet"
	botStr     = "Bot"
)

// RedirectStatsProcessor processes raw redirect statistics events,
// enriching them with geolocation and device information before storage.
type RedirectStatsProcessor struct {
	statsStorage domain.StatsEventAdder
	ipLocator    location.IPLocator
	logger       domain.Logger
}

// NewRedirectStatsProcessor creates a new RedirectStatsProcessor instance.
// Parameters:
//   - statsStorage: storage for persisting processed events
//   - logger: logger for recording warnings and errors
func NewRedirectStatsProcessor(statsStorage domain.StatsEventAdder, ipLocator location.IPLocator, logger domain.Logger) *RedirectStatsProcessor {
	return &RedirectStatsProcessor{
		statsStorage: statsStorage,
		ipLocator:    ipLocator,
		logger:       logger,
	}
}

// ProcessEvent processes a raw statistics event from JSON byte data.
// It parses the event, enriches it with geolocation and device type information,
// and stores the processed event.
//
// Returns an error if:
//   - JSON parsing fails
//   - Storage operation fails
func (rsp *RedirectStatsProcessor) ProcessEvent(ctx context.Context, eventData []byte) error {
	rawEvent, err := parseEvent(eventData)
	if err != nil {
		return err
	}

	processedEvent := rsp.convertEvent(rawEvent)
	err = rsp.statsStorage.AddStatsEvent(ctx, processedEvent)
	if err != nil {
		return err
	}

	return nil
}

func (rsp *RedirectStatsProcessor) convertEvent(event domain.RawStatsEvent) domain.ProcessedStatsEvent {
	processedEvent := domain.ProcessedStatsEvent{
		UrlToken:  event.UrlToken,
		Timestamp: event.Timestamp,
	}

	ipLocation, err := rsp.ipLocator.LocateIP(event.IP)
	if err != nil {
		rsp.logger.Warn("Failed to locate IP: " + err.Error())
		processedEvent.Country = unknowsStr
		processedEvent.City = unknowsStr
	} else {
		processedEvent.Country = ipLocation.Country
		processedEvent.City = ipLocation.City
	}

	processedEvent.DeviceType = getDeviceType(event.UserAgent)
	processedEvent.Referrer = event.Referrer

	return processedEvent
}

func parseEvent(eventData []byte) (domain.RawStatsEvent, error) {
	var event domain.RawStatsEvent
	err := json.Unmarshal(eventData, &event)
	if err != nil {
		return domain.RawStatsEvent{}, err
	}

	return event, nil
}

func getDeviceType(userAgent string) string {
	info := useragent.Parse(userAgent)

	if info.Mobile {
		return mobileStr
	} else if info.Tablet {
		return tabletStr
	} else if info.Desktop {
		return desktopStr
	} else if info.Bot {
		return botStr
	}

	return info.Name
}
