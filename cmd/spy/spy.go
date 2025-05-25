package main

import (
	"log"

	"github.com/natewong1313/spy/internal/db"
	"github.com/natewong1313/spy/internal/queries"
	"github.com/natewong1313/spy/scrapers/greenhouse"
)

func main() {
	db, err := db.New("postgres://user:password@127.0.0.1:5432/spydb?sslmode=disable")
	if err != nil {
		panic(err)
	}

	discoveryWorker := greenhouse.NewDiscoveryScraper()
	companies, err := discoveryWorker.Start()
	if err != nil {
		panic(err)
	}
	if err := queries.AddCompanies(companies, db); err != nil {
		panic(err)
	}
	log.Printf("added %d companies", len(companies))

	// scraper := greenhouse.New(mockCompany)
	// jobs, err := scraper.Start()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(jobs[0].Company)
	// err = queries.AddJobs(jobs, db)
	// if err != nil {
	// 	panic(err)
	// }

}
