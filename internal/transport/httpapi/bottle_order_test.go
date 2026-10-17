package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lpamplonadev/jadeacai-bkend/internal/domain"
)

func bottleOrderRequest(extraID string) string {
	return `{"customer":{"name":"Ana Silva","phone":"21999990000"},"acai":{"flavorId":"banana","sizeId":"size-300","bottleId":"bottle-acai","toppingIds":[],"sauceId":"none","condimentPositionId":"bottom","fruitIds":[],"extraIds":["` + extraID + `"]},"delivery":{"postalCode":"21770-000","street":"Rua Jade","number":"10","neighborhood":"Realengo"},"payment":{"method":"pix"},"estimatedTotalCents":1800}`
}

func newBottleOrderStore(enabled, extraEnabled bool) *testOrderStore {
	settings := domain.DefaultStoreSettings()
	open := true
	settings.ManualOverride = &open
	settings.BottleCatalog.Enabled = enabled
	settings.BottleCatalog.Price300Cents = 1800
	settings.BottleCatalog.Price500Cents = 2600
	settings.BottleCatalog.Extras[0].Enabled = extraEnabled
	settings.BottleCatalog.Extras[0].PriceCents = 250
	settings.DeliveryZones = []domain.DeliveryZone{{
		Name:          "Realengo",
		Neighborhoods: []string{"Realengo"},
		FeeCents:      300,
		Enabled:       true,
	}}
	return &testOrderStore{
		settings: settings,
		created:  createdOrder{ID: "bottle-order", OrderNumber: 1, OrderDate: "2026-10-14"},
	}
}

func TestCreateBottleOrderAllowsOnlyEnabledBottleAndExtras(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := newBottleOrderStore(true, true)
	store.catalog.Items = []catalogItemRecord{{
		ID: "size-300", Kind: string(domain.CatalogSize), Name: "300 ml", Available: true,
	}}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(bottleOrderRequest("leite-condensado")))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newRouter(store).ServeHTTP(response, request)

	if response.Code != http.StatusAccepted || !store.called {
		t.Fatalf("expected enabled bottle order to be accepted, got status %d: %s", response.Code, response.Body.String())
	}
	if store.order.Acai.BottleID != "bottle-acai" || store.order.Acai.ExtraIDs[0] != "leite-condensado" {
		t.Fatalf("expected bottle and optional extra to persist, got %+v", store.order.Acai)
	}
	if store.order.EstimatedTotalCents != 2350 || store.order.Acai.BottleID == "" {
		t.Fatalf("expected server total to equal bottle price + extra + delivery fee, got %d", store.order.EstimatedTotalCents)
	}
}

func TestCreateBottleOrderRejectsDisabledBottleAndExtra(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name         string
		bottleActive bool
		extraActive  bool
	}{
		{name: "bottle disabled", bottleActive: false, extraActive: true},
		{name: "extra disabled", bottleActive: true, extraActive: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := newBottleOrderStore(test.bottleActive, test.extraActive)
			request := httptest.NewRequest(http.MethodPost, "/api/v1/orders", strings.NewReader(bottleOrderRequest("leite-condensado")))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			newRouter(store).ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest || store.called {
				t.Fatalf("expected disabled bottle choice to be rejected before persistence, got status %d", response.Code)
			}
		})
	}
}
