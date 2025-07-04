package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	AddCompaniesQuery = `INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name, ashby_name)
	VALUES (@name, @platform_type, @platform_url, @created_at, @greenhouse_name, @ashby_name)
	ON CONFLICT (name) DO UPDATE SET 
		greenhouse_name=EXCLUDED.greenhouse_name, platform_type=EXCLUDED.platform_type,
		platform_url=EXCLUDED.platform_url, ashby_name=EXCLUDED.ashby_name;`
	GetCompaniesQuery                    = `SELECT * FROM company LIMIT $1 OFFSET $2;`
	GetPaginatedCompaniesByPlatformQuery = `SELECT * FROM company WHERE platform_type=$1 LIMIT $2 OFFSET $3;`
	NewCompanyQuery                      = "INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name, ashby_name) VALUES ($1, $2, $3, $4, $5, $6);"
	GetCompanyByNameQuery                = "SELECT * FROM company WHERE name=$1;"
)

func AddCompanies(companies []models.Company, db *pgxpool.Conn) error {
	batch := &pgx.Batch{}
	for _, company := range companies {
		args := pgx.NamedArgs{
			"name":            company.Name,
			"platform_type":   company.PlatformType,
			"platform_url":    company.PlatformURL,
			"created_at":      company.CreatedAt,
			"greenhouse_name": company.GreenhouseName,
			"ashby_name":      company.AshbyName,
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

func GetPaginatedCompanies(site string, page, limit int, db *pgxpool.Conn) (companies []models.Company, err error) {
	offset := (page - 1) * limit
	rows, err := db.Query(context.Background(), GetPaginatedCompaniesByPlatformQuery, site, limit, offset)
	if err != nil {
		return companies, err
	}
	defer rows.Close()
	for rows.Next() {
		company, err := scanCompany(rows)
		if err != nil {
			return companies, err
		}
		companies = append(companies, company)
	}
	return companies, nil
}

func NewCompany(company models.Company, db *pgxpool.Conn) error {
	_, err := db.Exec(context.Background(), NewCompanyQuery, company.Name, company.PlatformType, company.PlatformURL, company.CreatedAt, company.GreenhouseName, company.AshbyName)
	return err
}

func GetCompanyByName(name string, db *pgxpool.Conn) (models.Company, error) {
	row := db.QueryRow(context.Background(), GetCompanyByNameQuery, name)
	return scanCompany(row)
}

// utility function to colocate logic incase new fields are added to company
func scanCompany(row pgx.Row) (company models.Company, err error) {
	if err := row.Scan(&company.Name, &company.PlatformType, &company.PlatformURL, &company.CreatedAt, &company.GreenhouseName, &company.AshbyName); err != nil {
		return company, errors.Wrap(err, "query company")
	}
	return company, nil
}
