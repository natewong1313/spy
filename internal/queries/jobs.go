package queries

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	AddJobsQuery = `INSERT INTO job (url, company, title, locations, updated_at, created_at)
	VALUES (@url, @company, @title, @locations, @updated_at, @created_at)
	ON CONFLICT (url) DO UPDATE SET locations=EXCLUDED.locations, updated_at=EXCLUDED.updated_at;`
	DeleteOldJobsQuery = "DELETE FROM job WHERE company=$1 AND url !=ALL($2);"
)

// adds new jobs to the database, updates existing, and deletes invalid ones
func AddJobs(jobs []models.Job, db *pgxpool.Conn) error {
	if len(jobs) == 0 {
		return nil
	}
	company := jobs[0].Company

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
	for i := range jobs {
		_, err := results.Exec()
		if err != nil {
			results.Close()
			log.Printf("err: %s : %s : %s", jobs[i].URL, jobs[i].Company, jobs[i].Title)
			log.Println(jobs[i].Locations)
			return errors.Wrap(err, "execAddJobsQuery")
		}
	}
	results.Close()

	// delete jobs that werent found during the scrape, which are probably expired/filled jobs
	_, err := db.Exec(ctx, DeleteOldJobsQuery, company, jobURLs)
	if err != nil {
		return errors.Wrap(err, "deleteOldJobsQuery")
	}
	return nil
}
