package ashby

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natewong1313/spy/internal/models"
	"github.com/natewong1313/spy/internal/queries"
)

func Start(dbPool *pgxpool.Pool) {
	attrs := []slog.Attr{
		slog.String("site", "greenhouse"),
		slog.String("service", "master"),
	}
	handler := slog.NewTextHandler(os.Stdout, nil).WithAttrs(attrs)
	logger := slog.New(handler)

	// first search for any new companies
	companies, err := NewDiscoveryScraper().Start()
	if err != nil {
		logger.Error("discovery scraper err", slog.Any("err", err))
		// dont return for now
	}
	// we pool db connections since multiple workers are using the db
	dbConn, err := dbPool.Acquire(context.Background())
	if err != nil {
		logger.Error("acquire pool", slog.Any("err", err))
		return
	}
	defer dbConn.Release()

	// add the companies we just found
	if err := queries.AddCompanies(companies, dbConn); err != nil {
		logger.Error("add companies", slog.Any("err", err))
		return
	} else {
		logger.Info(fmt.Sprintf("added %d companies", len(companies)))
	}

	// work through companies in batches
	// lots of companies so we dont want to get all at once
	var allCompanies []models.Company
	page := 1
	limit := 50
	for {
		companies, err := queries.GetPaginatedCompanies("ashby", page, limit, dbConn)
		if err != nil {
			logger.Error("get companies", slog.Any("err", err))
			return
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
			logger.Error("unexpected err scraping jobs", slog.Any("err", err))
			continue
		}
		if err := queries.AddJobs(jobs, dbConn); err != nil {
			logger.Error("unexpected err adding jobs", slog.Any("err", err))
		}
	}
}
