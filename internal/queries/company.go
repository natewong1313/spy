package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	AddCompaniesQuery = `INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name)
	VALUES (@name, @platform_type, @platform_url, @created_at, @greenhouse_name)
	ON CONFLICT (name) DO UPDATE SET greenhouse_name=EXCLUDED.greenhouse_name, platform_type=EXCLUDED.platform_type, platform_url=EXCLUDED.platform_url;`
	NewCompanyQuery       = "INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name) VALUES ($1, $2, $3, $4, $5)"
	GetCompanyByNameQuery = "SELECT * FROM company WHERE name=$1"
)

func AddCompanies(companies []models.Company, db *pgx.Conn) error {
	batch := &pgx.Batch{}
	for _, company := range companies {
		args := pgx.NamedArgs{
			"name":            company.Name,
			"platform_type":   company.PlatformType,
			"platform_url":    company.PlatformURL,
			"created_at":      company.CreatedAt,
			"greenhouse_name": company.GreenhouseName,
		}
		batch.Queue(AddCompaniesQuery, args)
	}
	// lots of rows so we'll batch
	results := db.SendBatch(context.Background(), batch)
	defer results.Close()
	for range companies {
		_, err := results.Exec()
		if err != nil {
			return errors.Wrap(err, "execAddCompaniesQuery")
		}
	}
	return nil
}

func NewCompany(company models.Company, db *pgx.Conn) error {
	_, err := db.Exec(context.Background(), NewCompanyQuery, company.Name, company.PlatformType, company.PlatformURL, company.CreatedAt, company.GreenhouseName)
	return err
}

func GetCompanyByName(name string, db *pgx.Conn) (models.Company, error) {
	var company models.Company
	if err := db.QueryRow(context.Background(), GetCompanyByNameQuery, name).Scan(&company.Name, &company.PlatformType, &company.PlatformURL, &company.CreatedAt, &company.GreenhouseName); err != nil {
		return company, errors.Wrap(err, "query company by name")
	}
	return company, nil
}
