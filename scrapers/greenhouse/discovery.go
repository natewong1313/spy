package greenhouse

import (
	"fmt"
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

func New() *DiscoveryScraper {
	// jar, _ := cookiejar.New(nil)
	return &DiscoveryScraper{client: &http.Client{}, googleQueryURL: getQueryURL()}
}

func (ds *DiscoveryScraper) Start() error {
	_, err := ds.getGoogleSearchResults()
	if err != nil {
		return errors.Wrap(err, "getGoogleSearchResults")
	}
	return nil

}

func getQueryURL() string {
	weekAgo := time.Now().AddDate(0, 0, -7)
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
		// /url?q=https://boards.greenhouse.io/cialfo&sa=U&ved=2ahUKEwj27LPx-r6NAxUm48kDHd3uOXYQFnoECAMQAg&usg=AOvVaw0LIFj2cF-uYpWMcHetb8ei
		hrefURL, err := url.Parse(strings.Split(href, "/url?q=")[1])
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
		company := models.Company{
			Name:           companyName,
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
		return companies, nil
	} else {
		ds.googleQueryURL = "https://www.google.com" + nextURL
		// avoid rate limiting
		time.Sleep(5 * time.Second)
		return ds.getGoogleSearchResults()
	}
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
