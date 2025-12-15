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

type RedirectStatsProcessor struct {
	statsStorage domain.StatisticsStorage
	logger       domain.Logger
}

func NewRedirectStatsProcessor(statsStorage domain.StatisticsStorage, logger domain.Logger) *RedirectStatsProcessor {
	return &RedirectStatsProcessor{
		statsStorage: statsStorage,
		logger:       logger,
	}
}

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

	ipLocation, err := location.LocateIP(event.IP)
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

	return unknowsStr
}
