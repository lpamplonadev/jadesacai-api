package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lpamplonadev/jadeacai-bkend/internal/domain"
)

func TestAdminSettingsUpdatePreservesBottleCatalogWhenFieldIsOmitted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	apiKey := "test-admin-api-key-with-at-least-32-characters"
	t.Setenv("ADMIN_API_KEY", apiKey)
	settings := domain.DefaultStoreSettings()
	settings.BottleCatalog.Enabled = true
	settings.BottleCatalog.Price300Cents = 1800
	settings.BottleCatalog.Price500Cents = 2600
	store := &testOrderStore{settings: settings}
	requestBody := `{"weeklyHours":{"monday":{"enabled":false,"opensAt":"","closesAt":""},"tuesday":{"enabled":true,"opensAt":"19:00","closesAt":"23:00"},"wednesday":{"enabled":true,"opensAt":"19:00","closesAt":"23:00"},"thursday":{"enabled":true,"opensAt":"19:00","closesAt":"23:00"},"friday":{"enabled":true,"opensAt":"19:00","closesAt":"23:00"},"saturday":{"enabled":true,"opensAt":"17:00","closesAt":"23:00"},"sunday":{"enabled":true,"opensAt":"17:00","closesAt":"23:00"}},"story":{"title":"Nossa história","body":"Açaí feito com carinho."},"whatsAppNumber":"(21) 99999-0000","manualOverride":null}`
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/admin/settings", strings.NewReader(requestBody))
	request.Header.Set("Authorization", "Bearer "+apiKey)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newRouter(store).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
	if !store.settings.BottleCatalog.Enabled || store.settings.BottleCatalog.Price300Cents != 1800 || store.settings.BottleCatalog.Price500Cents != 2600 {
		t.Fatalf("expected omitted bottle settings to be preserved, got %+v", store.settings.BottleCatalog)
	}
}
