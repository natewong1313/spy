package ashby

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
	"github.com/natewong1313/spy/scrapers/shared"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type DiscoveryScraper struct {
	logger         *slog.Logger
	client         *http.Client
	googleQueryURL string
}

func NewDiscoveryScraper() *DiscoveryScraper {
	attrs := []slog.Attr{
		slog.String("site", "ashby"),
		slog.String("service", "discovery"),
	}
	handler := slog.NewTextHandler(os.Stdout, nil).WithAttrs(attrs)
	logger := slog.New(handler)

	return &DiscoveryScraper{logger: logger, client: &http.Client{}, googleQueryURL: getQueryURL()}
}

func (ds *DiscoveryScraper) Start() (totalCompanies []models.Company, err error) {
	ds.logger.Info("starting")
	for {
		if ds.googleQueryURL == "" {
			return totalCompanies, nil
		}
		companies, err := ds.getGoogleSearchResults()
		totalCompanies = append(totalCompanies, companies...)
		if err != nil {
			if err.Error() == "non 200 error code: 429" {
				ds.logger.Error("rate limited getting google results, sleeping")
				time.Sleep(30 * time.Second)
				continue
			}
			ds.logger.Error("getGoogleSearchResults", slog.Any("err", err))
			return totalCompanies, nil
		}
		ds.logger.Info(fmt.Sprintf("parsed %d companies", len(companies)))
		// rate limiting
		ds.logger.Debug("sleeping...")
		time.Sleep(30 * time.Second)
	}
}

func getQueryURL() string {
	// get results within x days
	weekAgo := time.Now().AddDate(0, 0, -1)
	return fmt.Sprintf("https://www.google.com/search?q=site:jobs.ashbyhq.com+after:%d-%02d-%d", weekAgo.Year(), weekAgo.Month(), weekAgo.Day())
}

func (ds *DiscoveryScraper) getGoogleSearchResults() (companies []models.Company, err error) {
	ds.logger.Info("fetching search results", slog.String("url", ds.googleQueryURL))

	resp, err := shared.DoGoogleSearchRequest(ds.googleQueryURL, ds.client)
	if err != nil {
		return companies, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return companies, errors.Wrap(err, "NewDocumentFromReader")
	}

	container := doc.Find("body div:nth-child(1)")
	searchResults := container.Find("a").Nodes

	// iterate through all search results
	for _, searchResult := range searchResults {
		// parse href
		var href string
		for _, attr := range searchResult.Attr {
			if attr.Key == "href" {
				href = attr.Val
				break
			}
		}
		if !strings.HasPrefix(href, "/url?q=https://jobs.ashbyhq.com") {
			continue
		}
		href = strings.Split(href, "/url?q=")[1]
		// /url?q=https://jobs.ashbyhq.com/deel&sa=U&ved=2ahUKEwjX74-d2MGNAxVMrlYBHdYBJa8QFnoECAoQAg&usg=AOvVaw0fUhN8Sn7WsaR0H_BOIUGy
		hrefURL, err := url.Parse(href)
		if err != nil {
			ds.logger.Error("unexpected url parse err", slog.Any("err", err), slog.String("url", href))
			continue
		}
		// /Silver/889efaa9-316e-48be-91ed-09d32962027d&sa=U&ved=2ahUKEwjX74-d2MGNAxVMrlYBHdYBJa8Q0gJ6BAgEEAU&usg=AOvVaw37rQ22obXsWEB8Qk86sTPc
		paths := strings.Split(hrefURL.Path, "/")
		if len(paths) < 1 {
			ds.logger.Error("unexpected path length", slog.String("path", hrefURL.Path))
			continue
		}
		// strip params, edge case for some urls
		companyName := strings.Split(paths[1], "&")[0]
		// no better way of doing this atm
		formattedCompanyName := cases.Title(language.English).String(companyName)
		ds.logger.Info("discovered " + formattedCompanyName)

		company := models.Company{
			Name:         formattedCompanyName,
			PlatformType: "ashby",
			PlatformURL:  "",
			CreatedAt:    time.Now(),
			AshbyName:    companyName,
		}
		companies = append(companies, company)
	}

	return companies, nil
}
