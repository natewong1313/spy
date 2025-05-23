package greenhouse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

type GreenhouseScraper struct {
	company models.Company
	client  *http.Client
}

func New(company models.Company) *GreenhouseScraper {

	return &GreenhouseScraper{company: company, client: &http.Client{}}
}

// should be ran as a go routine
func (gs *GreenhouseScraper) Start() error {
	log.Printf("starting scrape job for %s", gs.company.Name)
	departmentsResponse, err := gs.getDepartmentsData()
	if err != nil {
		return errors.Wrap(err, "getDepartmentsData")
	}

	jobs := gs.parseJobs(departmentsResponse)
	fmt.Println(len(jobs))

	return nil
}

// https://boards-api.greenhouse.io/v1/boards/{GREENHOUSE NAME}/departments/
func (gs *GreenhouseScraper) getDepartmentsData() (DepartmentsResponse, error) {
	var departments DepartmentsResponse

	url := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/departments/", gs.company.GreenhouseName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return departments, errors.Wrap(err, "build request")
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", gs.company.PlatformURL)
	req.Header.Set("Origin", gs.company.PlatformURL)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("If-None-Match", `W/"c83d52e7507679e97c702a72682c217f"`)
	req.Header.Set("Priority", "u=4")
	req.Header.Set("TE", "trailers")

	resp, err := gs.client.Do(req)
	if err != nil {
		return departments, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return departments, fmt.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return departments, errors.Wrap(err, "read body")
	}
	if err := json.Unmarshal(respBody, &departments); err != nil {
		return departments, errors.Wrap(err, "unmarshal body")
	}
	return departments, nil
}

// jobs that are scraped come categorized by department but we want to put them into one big list
func (gs *GreenhouseScraper) parseJobs(departments DepartmentsResponse) []models.Job {
	var parsedJobs []models.Job
	for _, department := range departments.Departments {
		for _, job := range department.Jobs {
			parsedJobs = append(parsedJobs, models.Job{
				// TODO: dont add id yet, for duplicate jobs
				Company:   job.CompanyName,
				Title:     job.Title,
				Locations: getLocationsFromRawJob(job),
				URL:       job.AbsoluteURL,
				UpdatedAt: job.UpdatedAt,
				CreatedAt: time.Now(),
			})
		}
	}
	return parsedJobs
}

func getLocationsFromRawJob(rawJob job) []string {
	// jobs can either be in metadata OR location field
	for _, metadata := range rawJob.Metadata {
		if metadata.Name == "Job Posting Location" {
			return metadata.Value.StringArr
		}
	}
	if rawJob.Location.Name != "" {
		return []string{rawJob.Location.Name}
	}
	log.Printf("no locations found for job: %s at %s", rawJob.Title, rawJob.CompanyName)
	return []string{}
}
