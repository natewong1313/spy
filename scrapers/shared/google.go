package shared

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

var PlatformsToURLs = map[string]string{
	"greenhouse": "https://boards.greenhouse.io",
	"ashby":      "https://jobs.ashbyhq.com",
}

type GoogleDiscoveryArgs struct {
	QueryURL     string
	PlatformType string // "greenhouse" | "ashby"
	Client       *http.Client
	ParseURLFunc func(url string) (string, string, error)
}

func DoGoogleCompanyDiscovery(args GoogleDiscoveryArgs) (companies []models.Company, nextQueryURL string, err error) {
	req, err := http.NewRequest("GET", args.QueryURL, nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "NewRequest")
	}
	req.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("accept-language", "en-US,en;q=0.5")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("priority", "u=0, i")
	req.Header.Set("sec-ch-ua", `"Chromium";v="136", "Brave";v="136", "Not.A/Brand";v="99"`)
	req.Header.Set("sec-ch-ua-arch", `"x86"`)
	req.Header.Set("sec-ch-ua-bitness", `"64"`)
	req.Header.Set("sec-ch-ua-full-version-list", `"Chromium";v="136.0.0.0", "Brave";v="136.0.0.0", "Not.A/Brand";v="99.0.0.0"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-model", `""`)
	req.Header.Set("sec-ch-ua-platform", `"Linux"`)
	req.Header.Set("sec-ch-ua-platform-version", `"6.14.6"`)
	req.Header.Set("sec-ch-ua-wow64", "?0")
	req.Header.Set("sec-fetch-dest", "document")
	req.Header.Set("sec-fetch-mode", "navigate")
	req.Header.Set("sec-fetch-site", "none")
	req.Header.Set("sec-fetch-user", "?1")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("upgrade-insecure-requests", "1")
	// req.Header.Set("user-agent", "Lynx/2.8.6rel.5 libwww-FM/2.14")
	req.Header.Set("User-Agent", "Links (2.29; Linux 6.11.0-13-generic x86_64; GNU C 13.2; text)")
	resp, err := args.Client.Do(req)
	if err != nil {
		return nil, "", errors.Wrap(err, "doRequest")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("non 200 error code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", errors.Wrap(err, "NewDocumentFromReader")
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
		baseURL := PlatformsToURLs[args.PlatformType]
		if !strings.HasPrefix(href, fmt.Sprintf("/url?q=%s", baseURL)) {
			continue
		}
		href = strings.Split(href, "/url?q=")[1]

		companyName, formattedCompanyName, err := args.ParseURLFunc(href)
		if err != nil {
			if err.Error() == "company previously discovered" {
				continue
			}
			return nil, "", err
		}
		company := models.Company{
			Name:         formattedCompanyName,
			PlatformType: args.PlatformType,
			PlatformURL:  "",
			CreatedAt:    time.Now(),
		}
		switch args.PlatformType {
		case "greenhouse":
			company.GreenhouseName = companyName
		case "ashby":
			company.AshbyName = companyName
		}
		companies = append(companies, company)
	}
	nextURL, err := parsePageButtons(doc)
	if err != nil {
		return nil, "", errors.Wrap(err, "parsePageButtons")
	}

	// end of results
	if nextURL == "#" {
		nextQueryURL = ""
	} else {
		nextQueryURL = "https://www.google.com" + nextURL
	}
	return companies, nextQueryURL, nil
}

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
