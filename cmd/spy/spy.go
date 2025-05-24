package main

import (
	"fmt"
	"time"

	"github.com/natewong1313/spy/internal/db"
	"github.com/natewong1313/spy/internal/models"
	"github.com/natewong1313/spy/internal/queries"
	"github.com/natewong1313/spy/scrapers/jobs/greenhouse"
)

func main() {
	db, err := db.New("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		panic(err)
	}
	mockCompany := models.Company{
		Name:           "Stripe",
		PlatformType:   "greenday",
		PlatformURL:    "https://stripe.com/",
		CreatedAt:      time.Now(),
		GreenhouseName: "stripe",
	}
	queries.NewCompany(mockCompany, db)
	scraper := greenhouse.New(mockCompany)
	jobs, err := scraper.Start()
	if err != nil {
		panic(err)
	}
	fmt.Println(jobs[0].Company)
	err = queries.AddJobs(jobs, db)
	if err != nil {
		panic(err)
	}

}
