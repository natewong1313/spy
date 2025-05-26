package shared

import (
	"fmt"
	"io"
	"net/http"

	"github.com/natewong1313/spy/internal/errors"
)

func DoGoogleSearchRequest(queryURL string, client *http.Client) (*http.Response, error) {
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "NewRequest")
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
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "doRequest")
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		resp.Body.Close()
		return nil, fmt.Errorf("non 200 error code: %d", resp.StatusCode)
	}
	return resp, nil
}

