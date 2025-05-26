package greenhouse

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natewong1313/spy/internal/models"
	"github.com/natewong1313/spy/internal/queries"
)

func Start(dbPool *pgxpool.Pool) {
	// first search for any new companies
	companies, err := NewDiscoveryScraper().Start()
	if err != nil {
		panic(err)
	}

	dbConn, err := dbPool.Acquire(context.Background())
	if err != nil {
		panic(err)
	}
	defer dbConn.Release()

	if err := queries.AddCompanies(companies, dbConn); err != nil {
		panic(err)
	}
	log.Printf("added %d companies", len(companies))

	// work through companies in batches
	var allCompanies []models.Company
	page := 1
	limit := 50
	for {
		companies, err := queries.GetPaginatedCompanies(page, limit, dbConn)
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
		jobs, err := NewJobsScraper(company).Start()
		if err != nil {
			log.Printf("unexpected error: %v", err)
			continue
		}
		if err := queries.AddJobs(jobs, dbConn); err != nil {
			log.Printf("unexpected error: %v", err)
		}
	}
}
