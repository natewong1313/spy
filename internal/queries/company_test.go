package queries

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/natewong1313/spy/internal/db"
	"github.com/natewong1313/spy/internal/models"
)

func TestCompanyQueries(t *testing.T) {
	dbPool, err := db.NewPool("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	defer dbPool.Close()
	dbConn, err := dbPool.Acquire(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	defer dbConn.Release()

	testCompany := models.Company{
		Name:           "test",
		PlatformType:   "greenhouse",
		PlatformURL:    ".",
		CreatedAt:      time.Now(),
		GreenhouseName: "stripe",
	}
	err = NewCompany(testCompany, dbConn)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	company, err := GetCompanyByName("stripe", dbConn)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(company)
}
