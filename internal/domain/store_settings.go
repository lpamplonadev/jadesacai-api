package domain

import (
	"errors"
	"strings"
	"time"
)

const (
	StoreSettingsRuleKey = "store_settings"
	StoreTimeZone        = "America/Sao_Paulo"
	DeliveryOrigin       = "Rua Nepomuceno, 12, Realengo, Rio de Janeiro - RJ"
)

var ErrInvalidStoreSettings = errors.New("invalid store settings")
var ErrStoreClosed = errors.New("store is closed")
var ErrDeliveryAreaUnavailable = errors.New("delivery area is unavailable")

type OperatingHours struct {
	Enabled  bool   `json:"enabled"`
	OpensAt  string `json:"opensAt"`
	ClosesAt string `json:"closesAt"`
}

type StoreStory struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type DeliveryZone struct {
	Name          string   `json:"name"`
	Neighborhoods []string `json:"neighborhoods"`
	FeeCents      int      `json:"feeCents"`
	Enabled       bool     `json:"enabled"`
}

type BottleExtraSetting struct {
	Key        string `json:"key"`
	Name       string `json:"name"`
	PriceCents int    `json:"priceCents"`
	Enabled    bool   `json:"enabled"`
}

type BottleProductSetting struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Enabled       bool   `json:"enabled"`
	Price300Cents int    `json:"price300Cents"`
	Price500Cents int    `json:"price500Cents"`
}

type BottleCatalogSettings struct {
	Enabled       bool                   `json:"enabled"`
	Price300Cents int                    `json:"price300Cents"`
	Price500Cents int                    `json:"price500Cents"`
	Bottles       []BottleProductSetting `json:"bottles"`
	Extras        []BottleExtraSetting   `json:"extras"`
}

type StoreSettings struct {
	WeeklyHours           map[string]OperatingHours `json:"weeklyHours"`
	Story                 StoreStory                `json:"story"`
	WhatsAppNumber        string                    `json:"whatsAppNumber"`
	ManualOverride        *bool                     `json:"manualOverride"`
	DeliveryOriginAddress string                    `json:"deliveryOriginAddress"`
	DeliveryZones         []DeliveryZone            `json:"deliveryZones"`
	BottleCatalog         BottleCatalogSettings     `json:"bottleCatalog"`
}

type StoreStatus struct {
	IsOpen         bool          `json:"isOpen"`
	ManualOverride *bool         `json:"manualOverride"`
	TimeZone       string        `json:"timeZone"`
	Settings       StoreSettings `json:"settings"`
}

var weekdayKeys = map[string]time.Weekday{
	"sunday":    time.Sunday,
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
}

var neighborhoodDiacritics = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a",
	"é", "e", "ê", "e", "í", "i",
	"ó", "o", "ô", "õ", "ú", "u", "ç", "c",
)

func NormalizeNeighborhood(value string) string {
	normalized := strings.ToLower(strings.Join(strings.Fields(value), " "))
	return neighborhoodDiacritics.Replace(normalized)
}

func FindDeliveryZone(settings StoreSettings, neighborhood string) (DeliveryZone, bool) {
	neighborhood = NormalizeNeighborhood(neighborhood)
	if neighborhood == "" {
		return DeliveryZone{}, false
	}
	for _, zone := range settings.DeliveryZones {
		if !zone.Enabled {
			continue
		}
		for _, coveredNeighborhood := range zone.Neighborhoods {
			if NormalizeNeighborhood(coveredNeighborhood) == neighborhood {
				return zone, true
			}
		}
	}
	return DeliveryZone{}, false
}

func BottleProducts(settings BottleCatalogSettings) []BottleProductSetting {
	if len(settings.Bottles) > 0 {
		return settings.Bottles
	}
	if settings.Enabled || settings.Price300Cents > 0 || settings.Price500Cents > 0 {
		return []BottleProductSetting{{
			ID:            "bottle-acai",
			Name:          "Açaí na Garrafa",
			Enabled:       settings.Enabled,
			Price300Cents: settings.Price300Cents,
			Price500Cents: settings.Price500Cents,
		}}
	}
	return []BottleProductSetting{}
}

func DefaultStoreSettings() StoreSettings {
	return StoreSettings{
		WeeklyHours: map[string]OperatingHours{
			"monday":    {},
			"tuesday":   {Enabled: true, OpensAt: "19:00", ClosesAt: "23:00"},
			"wednesday": {Enabled: true, OpensAt: "19:00", ClosesAt: "23:00"},
			"thursday":  {Enabled: true, OpensAt: "19:00", ClosesAt: "23:00"},
			"friday":    {Enabled: true, OpensAt: "19:00", ClosesAt: "23:00"},
			"saturday":  {Enabled: true, OpensAt: "17:00", ClosesAt: "23:00"},
			"sunday":    {Enabled: true, OpensAt: "17:00", ClosesAt: "23:00"},
		},
		Story: StoreStory{
			Title: "Um intervalo gostoso muda o dia.",
			Body:  "Açaí de verdade, feito com carinho em cada pedido.",
		},
		WhatsAppNumber:        "5521984848283",
		DeliveryOriginAddress: DeliveryOrigin,
		DeliveryZones:         []DeliveryZone{},
		BottleCatalog: BottleCatalogSettings{
			Bottles: []BottleProductSetting{},
			Extras: []BottleExtraSetting{
				{Key: "leite-condensado", Name: "Leite condensado"},
				{Key: "doce-de-leite", Name: "Doce de leite"},
				{Key: "creme-de-avela", Name: "Creme de avelã"},
			},
		},
	}
}

func ValidateStoreSettings(settings StoreSettings) error {
	if strings.TrimSpace(settings.Story.Title) == "" || len(settings.Story.Title) > 120 ||
		strings.TrimSpace(settings.Story.Body) == "" || len(settings.Story.Body) > 2000 {
		return ErrInvalidStoreSettings
	}

	whatsAppDigits := NormalizeBrazilianPhone(settings.WhatsAppNumber)
	if strings.HasPrefix(whatsAppDigits, "55") {
		whatsAppDigits = strings.TrimPrefix(whatsAppDigits, "55")
	}
	if !IsValidBrazilianMobilePhone(whatsAppDigits) {
		return ErrInvalidStoreSettings
	}
	if strings.TrimSpace(settings.DeliveryOriginAddress) == "" || len(settings.DeliveryOriginAddress) > 200 || len(settings.DeliveryZones) > 50 {
		return ErrInvalidStoreSettings
	}
	if settings.BottleCatalog.Price300Cents < 0 || settings.BottleCatalog.Price500Cents < 0 ||
		settings.BottleCatalog.Price300Cents > 100_000 || settings.BottleCatalog.Price500Cents > 100_000 ||
		(len(settings.BottleCatalog.Bottles) == 0 && settings.BottleCatalog.Enabled && (settings.BottleCatalog.Price300Cents == 0 || settings.BottleCatalog.Price500Cents == 0)) ||
		len(settings.BottleCatalog.Bottles) > 50 ||
		len(settings.BottleCatalog.Extras) != 3 {
		return ErrInvalidStoreSettings
	}
	bottleIDs := make(map[string]struct{}, len(settings.BottleCatalog.Bottles))
	for _, bottle := range settings.BottleCatalog.Bottles {
		id := strings.TrimSpace(bottle.ID)
		if id == "" || len(id) > 100 || strings.TrimSpace(bottle.Name) == "" || len(bottle.Name) > 80 ||
			bottle.Price300Cents < 0 || bottle.Price300Cents > 100_000 || bottle.Price500Cents < 0 || bottle.Price500Cents > 100_000 ||
			(bottle.Enabled && (bottle.Price300Cents == 0 || bottle.Price500Cents == 0)) {
			return ErrInvalidStoreSettings
		}
		if _, exists := bottleIDs[id]; exists {
			return ErrInvalidStoreSettings
		}
		bottleIDs[id] = struct{}{}
	}
	bottleExtraKeys := make(map[string]struct{}, len(settings.BottleCatalog.Extras))
	for _, extra := range settings.BottleCatalog.Extras {
		key := strings.TrimSpace(extra.Key)
		if key == "" || strings.TrimSpace(extra.Name) == "" || len(extra.Name) > 80 || extra.PriceCents < 0 || extra.PriceCents > 100_000 {
			return ErrInvalidStoreSettings
		}
		if _, exists := bottleExtraKeys[key]; exists {
			return ErrInvalidStoreSettings
		}
		bottleExtraKeys[key] = struct{}{}
	}
	zoneNames := make(map[string]struct{}, len(settings.DeliveryZones))
	neighborhoods := make(map[string]struct{})
	for _, zone := range settings.DeliveryZones {
		zoneName := NormalizeNeighborhood(zone.Name)
		if zoneName == "" || len(zone.Name) > 80 || zone.FeeCents < 0 || zone.FeeCents > 100_000 ||
			len(zone.Neighborhoods) == 0 || len(zone.Neighborhoods) > 50 {
			return ErrInvalidStoreSettings
		}
		if _, exists := zoneNames[zoneName]; exists {
			return ErrInvalidStoreSettings
		}
		zoneNames[zoneName] = struct{}{}
		for _, neighborhood := range zone.Neighborhoods {
			name := NormalizeNeighborhood(neighborhood)
			if name == "" || len(neighborhood) > 80 {
				return ErrInvalidStoreSettings
			}
			if _, exists := neighborhoods[name]; exists {
				return ErrInvalidStoreSettings
			}
			neighborhoods[name] = struct{}{}
		}
	}

	for day, hours := range settings.WeeklyHours {
		if _, validDay := weekdayKeys[day]; !validDay {
			return ErrInvalidStoreSettings
		}
		if !hours.Enabled {
			continue
		}
		opensAt, openErr := time.Parse("15:04", hours.OpensAt)
		closesAt, closeErr := time.Parse("15:04", hours.ClosesAt)
		if openErr != nil || closeErr != nil || !opensAt.Before(closesAt) {
			return ErrInvalidStoreSettings
		}
	}
	return nil
}

func StoreIsOpenAt(settings StoreSettings, now time.Time) bool {
	if settings.ManualOverride != nil {
		return *settings.ManualOverride
	}

	location, err := time.LoadLocation(StoreTimeZone)
	if err != nil {
		return false
	}
	localTime := now.In(location)
	var dayKey string
	for key, weekday := range weekdayKeys {
		if weekday == localTime.Weekday() {
			dayKey = key
			break
		}
	}

	hours, exists := settings.WeeklyHours[dayKey]
	if !exists || !hours.Enabled {
		return false
	}
	currentTime := localTime.Format("15:04")
	return currentTime >= hours.OpensAt && currentTime < hours.ClosesAt
}

func NewStoreStatus(settings StoreSettings, now time.Time) StoreStatus {
	return StoreStatus{
		IsOpen:         StoreIsOpenAt(settings, now),
		ManualOverride: settings.ManualOverride,
		TimeZone:       StoreTimeZone,
		Settings:       settings,
	}
}
