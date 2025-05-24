package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	AddJobsQuery = `INSERT INTO job (url, company, title, locations, updated_at, created_at)
	VALUES (@url, @company, @title, @locations, @updated_at, @created_at)
	ON CONFLICT (url) DO UPDATE SET locations=EXCLUDED.locations, updated_at=EXCLUDED.updated_at;`
	DeleteOldJobsQuery = "DELETE FROM job WHERE url !=ALL($1);"
)

// adds new jobs to the database, updates existing, and deletes invalid ones
func AddJobs(jobs []models.Job, db *pgx.Conn) error {
	// store jobURLs during this loop for deletion purposes
	jobURLs := make([]string, len(jobs))
	batch := &pgx.Batch{}
	for _, job := range jobs {
		jobURLs = append(jobURLs, job.URL)
		args := pgx.NamedArgs{
			"url":        job.URL,
			"company":    job.Company,
			"title":      job.Title,
			"locations":  job.Locations,
			"updated_at": job.UpdatedAt,
			"created_at": job.CreatedAt,
		}
		batch.Queue(AddJobsQuery, args)
	}

	ctx := context.Background()
	// lots of rows so we'll batch
	results := db.SendBatch(ctx, batch)
	for range jobs {
		_, err := results.Exec()
		if err != nil {
			results.Close()
			return errors.Wrap(err, "execAddJobsQuery")
		}
	}
	results.Close()

	// delete jobs that werent found during the scrape, which are probably expired/filled jobs
	_, err := db.Exec(ctx, DeleteOldJobsQuery, jobURLs)
	if err != nil {
		return errors.Wrap(err, "deleteOldJobsQuery")
	}
	return nil
}
