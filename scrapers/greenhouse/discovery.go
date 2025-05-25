package greenhouse

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/natewong1313/spy/internal/errors"
	"github.com/natewong1313/spy/internal/models"
)

type DiscoveryScraper struct {
	client         *http.Client
	googleQueryURL string
}

func NewDiscoveryScraper() *DiscoveryScraper {
	// jar, _ := cookiejar.New(nil)
	return &DiscoveryScraper{client: &http.Client{}, googleQueryURL: getQueryURL()}
}

func (ds *DiscoveryScraper) Start() ([]models.Company, error) {
	// originally was going to be recursive but it makes more sense to update db as we get results
	// _, err := ds.getGoogleSearchResults()
	// if err != nil {
	// 	return errors.Wrap(err, "getGoogleSearchResults")
	// }
	// return nil

	totalCompanies := []models.Company{}
	for {
		if ds.googleQueryURL == "" {
			return totalCompanies, nil
		}
		companies, err := ds.getGoogleSearchResults()
		totalCompanies = append(totalCompanies, companies...)
		if err != nil {
			log.Printf("%v", err)
			return totalCompanies, nil
		}
		log.Printf("parsed %d companies", len(companies))

		// rate limiting
		time.Sleep(5)
	}
}

func getQueryURL() string {
	// get results within x days
	weekAgo := time.Now().AddDate(0, 0, -1)
	return fmt.Sprintf("https://www.google.com/search?q=site:boards.greenhouse.io+after:%d-%02d-%d", weekAgo.Year(), weekAgo.Month(), weekAgo.Day())
}

func (ds *DiscoveryScraper) getChallengePage() error {
	req, err := http.NewRequest("GET", ds.googleQueryURL, nil)
	if err != nil {
		return errors.Wrap(err, "NewRequest")
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-GPC", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://www.google.com/search?client=firefox-b-1-d&q=site%3Aboards.greenhouse.io+after%3A2025-05-17")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")
	resp, err := ds.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "DoRequest")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("non 200 error code: %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return errors.Wrap(err, "NewDocumentFromReader")
	}
	_ = doc
	// nodes := doc.Find("a:contains('click here')").Nodes
	// if len(nodes) != 1 {
	// 	return fmt.Errorf("invalid num of nodes: %d", len(nodes))
	// }
	// var href string
	// for _, attr := range nodes[0].Attr {
	// 	if attr.Key == "href" {
	// 		href = attr.Val
	// 		break
	// 	}
	// }

	return err
}

func (ds *DiscoveryScraper) getGoogleSearchResults() ([]models.Company, error) {
	var companies []models.Company

	log.Printf("requesting %s", ds.googleQueryURL)
	req, err := http.NewRequest("GET", ds.googleQueryURL, nil)
	if err != nil {
		return companies, errors.Wrap(err, "NewRequest")
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
	resp, err := ds.client.Do(req)
	if err != nil {
		return companies, errors.Wrap(err, "doRequest")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return companies, fmt.Errorf("non 200 error code: %d", resp.StatusCode)
	}

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
			log.Printf("unexpected parse err: %v for url %s", err, href)
			continue
		}
		// /coupang/jobs/6536235&sa=U&ved=2ahUKEwiIyszA_76NAxXQgFYBHfs1CTwQFnoECAoQAg&usg=AOvVaw33QLvcRWHJXhHWdKAkF_Ii
		paths := strings.Split(hrefURL.Path, "/")
		if len(paths) < 1 {
			log.Printf("unexpected path length: %s", hrefURL.Path)
			continue
		}
		// strip params, edge case for some urls
		companyName := strings.Split(paths[1], "&")[0]
		var formattedCompanyName string
		if companyName == "embed" {
			decodedURL, _ := url.QueryUnescape(href)
			companyName, formattedCompanyName, err = ds.getEmbedListingDetails(decodedURL)
			if err != nil {
				log.Printf("getEmbedListingDetails err: %v", err)
				continue
			}
		} else {
			formattedCompanyName, err = getFormattedCompanyName(companyName)
			if err != nil {
				log.Printf("getFormattedCompanyName err: %v", err)
				continue
			}
		}

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
		return companies, err
	}

	// end of results
	if nextURL == "#" {
		ds.googleQueryURL = ""
	} else {
		ds.googleQueryURL = "https://www.google.com" + nextURL
	}
	return companies, nil

	// // end of results
	// if nextURL == "#" {
	// 	return companies, nil
	// } else {
	// 	ds.googleQueryURL = "https://www.google.com" + nextURL
	// 	// avoid rate limiting
	// 	time.Sleep(5 * time.Second)
	// 	return ds.getGoogleSearchResults()
	// }
}

// some search results show up as embeds so we need to go on the page itself to get the company name
func (ds *DiscoveryScraper) getEmbedListingDetails(embedURL string) (string, string, error) {
	log.Printf("embed url: %s", embedURL)
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:138.0) Gecko/20100101 Firefox/138.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	// req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
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
	parsedJobURL, err := url.Parse(actualJobURL)
	if err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("parsing %s", actualJobURL))
	}
	paths := strings.Split(parsedJobURL.Path, "/")
	if len(paths) < 2 {
		return "", "", fmt.Errorf("failed to parse path: %s", parsedJobURL.Path)
	}
	companyName := paths[1]

	var appDetails embedAppJSONResponse
	rawAppJSON := doc.Find("script[type='application/ld+json']").Text()
	if err := json.Unmarshal([]byte(rawAppJSON), &appDetails); err != nil {
		return "", "", errors.Wrap(err, "get app json")
	}
	formattedCompanyName := appDetails.HiringOrganization.Name

	return companyName, formattedCompanyName, nil

}

// get the fancy formatted company name via an api call
func getFormattedCompanyName(companyName string) (string, error) {
	log.Printf("getting formatted name for %s", companyName)
	resp, err := http.Get("https://boards-api.greenhouse.io/v1/boards/" + companyName)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("get company name: non-200 error code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "read company body")
	}
	var responseBody companyNameResponse
	if err := json.Unmarshal(body, &responseBody); err != nil {
		return "", err
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
