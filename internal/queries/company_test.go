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
	q := New(db)
	testCompany := models.Company{
		Name:           "test",
		PlatformType:   "greenhouse",
		PlatformURL:    ".",
		CreatedAt:      time.Now(),
		GreenhouseName: "stripe",
	}
	err = q.NewCompany(testCompany)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	company, err := q.GetCompanyByName("stripe")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(company)
}
