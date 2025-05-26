package ashby

import "time"

type ashbyJobApiResponse struct {
	Jobs []struct {
		ID                 string              `json:"id"`
		Title              string              `json:"title"`
		Department         string              `json:"department"`
		Team               string              `json:"team"`
		EmploymentType     string              `json:"employmentType"`
		Location           string              `json:"location"`
		SecondaryLocations []secondaryLocation `json:"secondaryLocations"`
		PublishedAt        time.Time           `json:"publishedAt"`
		IsListed           bool                `json:"isListed"`
		IsRemote           bool                `json:"isRemote"`
		Address            struct {
			PostalAddress struct {
				AddressCountry  string `json:"addressCountry"`
				AddressLocality string `json:"addressLocality"`
			} `json:"postalAddress"`
		} `json:"address"`
		JobURL           string `json:"jobUrl"`
		ApplyURL         string `json:"applyUrl"`
		DescriptionHTML  string `json:"descriptionHtml"`
		DescriptionPlain string `json:"descriptionPlain"`
	} `json:"jobs"`
	APIVersion string `json:"apiVersion"`
}

type secondaryLocation struct {
	Location string `json:"location"`
}
