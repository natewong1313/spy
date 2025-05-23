package queries

import (
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

const (
	NewCompanyQuery       = "INSERT INTO company (name, platform_type, platform_url, created_at, greenhouse_name) VALUES ($1, $2, $3, $4, $5)"
	GetCompanyByNameQuery = "SELECT * FROM company WHERE name=$1"
)

func (q *QueryEngine) NewCompany(company models.Company) error {
	_, err := q.db.Exec(NewCompanyQuery, company.Name, company.PlatformType, company.PlatformURL, company.CreatedAt, company.GreenhouseName)
	return err
}

func (q *QueryEngine) GetCompanyByName(name string) (models.Company, error) {
	var company models.Company
	if err := q.db.QueryRow(GetCompanyByNameQuery, name).Scan(&company.Name, &company.PlatformType, &company.PlatformURL, &company.CreatedAt, &company.GreenhouseName); err != nil {
		return company, errors.Wrap(err, "query company by name")
	}
	return company, nil
}
