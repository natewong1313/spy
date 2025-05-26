package greenhouse

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
	"github.com/natewong1313/spy/scrapers/shared"
)

type DiscoveryScraper struct {
	logger         *slog.Logger
	client         *http.Client
	googleQueryURL string
}

func NewDiscoveryScraper() *DiscoveryScraper {
	attrs := []slog.Attr{
		slog.String("site", "greenhouse"),
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
			ds.logger.Error("getGoogleSearchResults", slog.Any("err", err))
			return totalCompanies, nil
		}
		ds.logger.Info(fmt.Sprintf("parsed %d companies", len(companies)))
		// rate limiting
		ds.logger.Debug("sleeping...")
		time.Sleep(5 * time.Second)
	}
}

func getQueryURL() string {
	// get results within x days
	weekAgo := time.Now().AddDate(0, 0, -1)
	return fmt.Sprintf("https://www.google.com/search?q=site:boards.greenhouse.io+after:%d-%02d-%d", weekAgo.Year(), weekAgo.Month(), weekAgo.Day())
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
	pages := container.Find("a").Nodes

	for _, page := range pages {
		// parse href
		var href string
		for _, attr := range page.Attr {
			if attr.Key == "href" {
				href = attr.Val
				break
			}
		}
		if !strings.HasPrefix(href, "/url?q=https://boards.greenhouse.io") {
			continue
		}
		href = strings.Split(href, "/url?q=")[1]
		// /url?q=https://boards.greenhouse.io/cialfo&sa=U&ved=2ahUKEwj27LPx-r6NAxUm48kDHd3uOXYQFnoECAMQAg&usg=AOvVaw0LIFj2cF-uYpWMcHetb8ei
		hrefURL, err := url.Parse(href)
		if err != nil {
			ds.logger.Error("unexpected url parse err", slog.Any("err", err), slog.String("url", href))
			continue
		}
		// /coupang/jobs/6536235&sa=U&ved=2ahUKEwiIyszA_76NAxXQgFYBHfs1CTwQFnoECAoQAg&usg=AOvVaw33QLvcRWHJXhHWdKAkF_Ii
		paths := strings.Split(hrefURL.Path, "/")
		if len(paths) < 1 {
			ds.logger.Error("unexpected path length", slog.String("path", hrefURL.Path))
			continue
		}
		// strip params, edge case for some urls
		companyName := strings.Split(paths[1], "&")[0]
		var formattedCompanyName string
		if companyName == "embed" {
			decodedURL, _ := url.QueryUnescape(href)
			companyName, formattedCompanyName, err = ds.getEmbedListingDetails(decodedURL)
			if err != nil {
				ds.logger.Error("getEmbedListingDetails", slog.Any("err", err))
				continue
			}
		} else {
			formattedCompanyName, err = ds.getFormattedCompanyName(companyName)
			if err != nil {
				ds.logger.Error("getFormattedCompanyName", slog.Any("err", err))
				continue
			}
		}
		ds.logger.Info("discovered " + formattedCompanyName)

		company := models.Company{
			Name:           formattedCompanyName,
			PlatformType:   "greenhouse",
			PlatformURL:    "",
			CreatedAt:      time.Now(),
			GreenhouseName: companyName,
		}
		companies = append(companies, company)
	}
	nextURL, err := parsePageButtons(doc)
	if err != nil {
		return companies, errors.Wrap(err, "parsePageButtons")
	}

	// end of results
	if nextURL == "#" {
		ds.googleQueryURL = ""
	} else {
		ds.googleQueryURL = "https://www.google.com" + nextURL
	}
	return companies, nil
}

// some search results show up as embeds so we need to go on the page itself to get the company name
func (ds *DiscoveryScraper) getEmbedListingDetails(embedURL string) (string, string, error) {
	ds.logger.Info("getting embed listing details", slog.String("url", embedURL))
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", "https://www.google.com/")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")

	resp, err := ds.client.Do(req)
	if err != nil {
		return "", "", errors.Wrap(err, "doRequest")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", "", fmt.Errorf("non-200 error code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "NewDocumentFromReader")
	}

	// https://boards.greenhouse.io/wizardcommerce/jobs/5545455004
	actualJobURL, found := doc.Find("meta[property='og:url']").Attr("content")
	if !found {
		return "", "", fmt.Errorf("embed job url not found")
	}
	// convert the url to a *url.URL so we can get the path
	parsedJobURL, err := url.Parse(actualJobURL)
	if err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("parsing %s", actualJobURL))
	}
	paths := strings.Split(parsedJobURL.Path, "/")
	if len(paths) < 2 {
		return "", "", fmt.Errorf("failed to parse path: %s", parsedJobURL.Path)
	}
	companyName := paths[1]

	// greenhouse stores some information about the company in a json object in the page
	var appDetails embedAppJSONResponse
	rawAppJSON := doc.Find("script[type='application/ld+json']").Text()
	if err := json.Unmarshal([]byte(rawAppJSON), &appDetails); err != nil {
		return "", "", errors.Wrap(err, "get app json")
	}
	formattedCompanyName := appDetails.HiringOrganization.Name

	return companyName, formattedCompanyName, nil

}

// get the fancy formatted company name via an api call
func (ds *DiscoveryScraper) getFormattedCompanyName(companyName string) (string, error) {
	ds.logger.Info("getting formatted company name", slog.String("company", companyName))
	resp, err := http.Get("https://boards-api.greenhouse.io/v1/boards/" + companyName)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 error code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read company body")
	}
	var responseBody companyNameResponse
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return "", errors.Wrap(err, "unmarshal json")
	}
	return responseBody.Name, nil
}

// check if we can navigate to another page
func parsePageButtons(doc *goquery.Document) (string, error) {
	children := doc.Find("body").Children()
	table := children.Closest("table")
	td := table.Find("td")
	switch len(td.Nodes) {
	// first page
	case 1:
		nextURL, found := td.Find("a").Attr("href")
		if !found {
			return "", fmt.Errorf("failed to find next page url on first page")
		}
		return nextURL, nil
	// back and forward buttons
	case 5:
		nextURL, found := table.Find("td:nth-child(4)").Find("a").Attr("href")
		if !found {
			return "", fmt.Errorf("failed to find next page url")
		}
		return nextURL, nil

	}
	return "", fmt.Errorf("unknown nodes length: %d", len(td.Nodes))
}
