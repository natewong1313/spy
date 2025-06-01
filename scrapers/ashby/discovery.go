package ashby

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

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

func (ds *DiscoveryScraper) parseURLFromGoogle(rawURL string) (string, string, error) {
	// /url?q=https://jobs.ashbyhq.com/deel&sa=U&ved=2ahUKEwjX74-d2MGNAxVMrlYBHdYBJa8QFnoECAoQAg&usg=AOvVaw0fUhN8Sn7WsaR0H_BOIUGy
	hrefURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("unexpected url parse err for url %s", rawURL))
	}
	// /Silver/889efaa9-316e-48be-91ed-09d32962027d&sa=U&ved=2ahUKEwjX74-d2MGNAxVMrlYBHdYBJa8Q0gJ6BAgEEAU&usg=AOvVaw37rQ22obXsWEB8Qk86sTPc
	paths := strings.Split(hrefURL.Path, "/")
	if len(paths) < 1 {
		return "", "", fmt.Errorf("unexpected path length for %s", hrefURL.Path)
	}
	// strip params, edge case for some urls
	companyName := strings.Split(paths[1], "&")[0]
	companyName, _ = url.QueryUnescape(companyName)
	// no better way of doing this atm
	formattedCompanyName := cases.Title(language.English).String(companyName)
	ds.logger.Info("discovered " + formattedCompanyName)
	return companyName, formattedCompanyName, nil
}

func (ds *DiscoveryScraper) getGoogleSearchResults() (companies []models.Company, err error) {
	ds.logger.Info("fetching search results", slog.String("url", ds.googleQueryURL))

	args := shared.GoogleDiscoveryArgs{
		QueryURL:     ds.googleQueryURL,
		PlatformType: "ashby",
		Client:       ds.client,
		ParseURLFunc: ds.parseURLFromGoogle,
	}
	companies, nextQueryURL, err := shared.DoGoogleCompanyDiscovery(args)
	ds.googleQueryURL = nextQueryURL
	return companies, err
}
