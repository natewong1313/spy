package main

import (
	"log"

	"github.com/natewong1313/spy/internal/db"
	"github.com/natewong1313/spy/internal/models"
	"github.com/natewong1313/spy/internal/queries"
	"github.com/natewong1313/spy/scrapers/greenhouse"
)

func main() {
	db, err := db.New("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	// first we find new companies
	discoveryWorker := greenhouse.NewDiscoveryScraper()
	companies, err := discoveryWorker.Start()
	if err != nil {
		panic(err)
	}
	if err := queries.AddCompanies(companies, db); err != nil {
		panic(err)
	}
	log.Printf("added %d companies", len(companies))

	// work through companies in batches
	var allCompanies []models.Company
	page := 1
	limit := 50
	for {
		companies, err := queries.GetPaginatedCompanies(page, limit, db)
		if err != nil {
			panic(err)
		}
		allCompanies = append(allCompanies, companies...)
		if len(companies) < limit {
			break
		}
		page++
	}

	for _, company := range allCompanies {
		jobs, err := greenhouse.NewJobsScraper(company).Start()
		if err != nil {
			log.Printf("unexpected error: %v", err)
			continue
		}
		if err := queries.AddJobs(jobs, db); err != nil {
			log.Printf("unexpected error: %v", err)
		}
	}

}
