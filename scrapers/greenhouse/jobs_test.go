package greenhouse

import (
	"testing"
	"time"

	"github.com/natewong1313/spy/internal/models"
)

func TestGreenhouse(t *testing.T) {
	mockCompany := models.Company{
		Name:           "Stripe",
		PlatformType:   "greenday",
		PlatformURL:    "https://stripe.com/",
		CreatedAt:      time.Now(),
		GreenhouseName: "stripe",
	}
	scraper := NewJobsScraper(mockCompany)
	jobs, err := scraper.Start()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	_ = jobs
}
