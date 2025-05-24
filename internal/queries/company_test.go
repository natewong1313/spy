package queries

import (
	"fmt"
	"testing"
	"time"

	"github.com/natewong1313/spy/internal/db"
	"github.com/natewong1313/spy/internal/models"
)

func TestCompanyQueries(t *testing.T) {
	db, err := db.New("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	testCompany := models.Company{
		Name:           "test",
		PlatformType:   "greenhouse",
		PlatformURL:    ".",
		CreatedAt:      time.Now(),
		GreenhouseName: "stripe",
	}
	err = NewCompany(testCompany, db)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	company, err := GetCompanyByName("stripe", db)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(company)
}
