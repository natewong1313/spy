package greenhouse

import "testing"

func TestDiscovery(t *testing.T) {
	scraper := New()
	err := scraper.Start()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
