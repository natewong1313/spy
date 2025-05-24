package queries

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	NewCompanyQuery       = "INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name) VALUES ($1, $2, $3, $4, $5)"
	GetCompanyByNameQuery = "SELECT * FROM company WHERE name=$1"
)

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
