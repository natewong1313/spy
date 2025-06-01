package ashby

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type JobsScraper struct {
	logger  *slog.Logger
	company models.Company
	client  *http.Client
}

func NewJobsScraper(company models.Company) *JobsScraper {
	attrs := []slog.Attr{
		slog.String("site", "ashby"),
		slog.String("service", "jobs"),
		slog.String("company", company.Name),
	}
	handler := slog.NewTextHandler(os.Stdout, nil).WithAttrs(attrs)
	logger := slog.New(handler)

	return &JobsScraper{logger: logger, company: company, client: &http.Client{}}
}

// should be ran as a go routine
func (js *JobsScraper) Start() (jobs []models.Job, err error) {
	js.logger.Info("starting scrape job")
	jobs, err = js.getJobsData()
	if err != nil {
		return jobs, errors.Wrap(err, "getDepartmentsData")
	}
	return jobs, nil
}

func (js *JobsScraper) getJobsData() (jobs []models.Job, err error) {
	fmt.Println(js.company.AshbyName)
	req, err := http.NewRequest("GET", "https://api.ashbyhq.com/posting-api/job-board/"+js.company.AshbyName, nil)
	if err != nil {
		return jobs, errors.Wrap(err, "build request")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := js.client.Do(req)
	if err != nil {
		return jobs, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return jobs, fmt.Errorf("non 200 status code: %d", resp.StatusCode)
	}
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return jobs, errors.Wrap(err, "read body")
	}

	var apiResponse ashbyJobApiResponse
	if err := json.Unmarshal(bodyText, &apiResponse); err != nil {
		return jobs, errors.Wrap(err, "unmarshal api response")
	}

	for _, job := range apiResponse.Jobs {
		parsedJob := models.Job{
			URL:       job.JobURL,
			Company:   cases.Title(language.English).String(js.company.Name),
			Title:     job.Title,
			UpdatedAt: job.PublishedAt,
			CreatedAt: time.Now(),
		}
		locations := []string{job.Location}
		for _, location := range job.SecondaryLocations {
			locations = append(locations, location.Location)
		}
		parsedJob.Locations = locations
		jobs = append(jobs, parsedJob)
	}
	return jobs, nil
}
