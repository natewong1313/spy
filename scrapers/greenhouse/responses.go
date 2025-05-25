package greenhouse

type companyNameResponse struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type embedAppJSONResponse struct {
	Context            string `json:"@context"`
	Type               string `json:"@type"`
	HiringOrganization struct {
		Type string `json:"@type"`
		Name string `json:"name"`
	} `json:"hiringOrganization"`
	Title       string `json:"title"`
	DatePosted  string `json:"datePosted"`
	JobLocation struct {
		Type    string `json:"@type"`
		Address struct {
			Type            string `json:"@type"`
			AddressLocality string `json:"addressLocality"`
			AddressRegion   string `json:"addressRegion"`
			AddressCountry  any    `json:"addressCountry"`
			PostalCode      any    `json:"postalCode"`
		} `json:"address"`
	} `json:"jobLocation"`
	Description string `json:"description"`
}
