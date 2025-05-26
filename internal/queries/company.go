package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	AddCompaniesQuery = `INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name)
	VALUES (@name, @platform_type, @platform_url, @created_at, @greenhouse_name)
	ON CONFLICT (name) DO UPDATE SET greenhouse_name=EXCLUDED.greenhouse_name, platform_type=EXCLUDED.platform_type, platform_url=EXCLUDED.platform_url;`
	GetCompaniesQuery           = `SELECT * FROM company LIMIT $1 OFFSET $2;`
	GetGreenhouseCompaniesQuery = `SELECT * FROM company WHERE platform_type='greenhouse' LIMIT $1 OFFSET $2;`
	NewCompanyQuery             = "INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name) VALUES ($1, $2, $3, $4, $5);"
	GetCompanyByNameQuery       = "SELECT * FROM company WHERE name=$1;"
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

func GetPaginatedGreenhouseCompanies(page, limit int, db *pgxpool.Conn) ([]models.Company, error) {
	var companies []models.Company
	offset := (page - 1) * limit
	rows, err := db.Query(context.Background(), GetGreenhouseCompaniesQuery, limit, offset)
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

func GetPaginatedCompanies(page, limit int, db *pgxpool.Conn) ([]models.Company, error) {
	var companies []models.Company
	offset := (page - 1) * limit
	rows, err := db.Query(context.Background(), GetCompaniesQuery, limit, offset)
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
	_, err := db.Exec(context.Background(), NewCompanyQuery, company.Name, company.PlatformType, company.PlatformURL, company.CreatedAt, company.GreenhouseName)
	return err
}

func GetCompanyByName(name string, db *pgxpool.Conn) (models.Company, error) {
	row := db.QueryRow(context.Background(), GetCompanyByNameQuery, name)
	return scanCompany(row)
}

// utility function to colocate logic incase new fields are added to company
func scanCompany(row pgx.Row) (models.Company, error) {
	var company models.Company
	if err := row.Scan(&company.Name, &company.PlatformType, &company.PlatformURL, &company.CreatedAt, &company.GreenhouseName); err != nil {
		return company, errors.Wrap(err, "query company")
	}
	return company, nil
}
