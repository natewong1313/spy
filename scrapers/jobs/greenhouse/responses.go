// contains structs for various responses
package greenhouse

import (
	"encoding/json"
	"fmt"
	"time"
)

type DepartmentsResponse struct {
	Departments []department `json:"departments"`
}

type department struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parent_id"`
	ChildIDs []int  `json:"child_ids"`
	Jobs     []job  `json:"jobs"`
}

type job struct {
	ID             int        `json:"id"`
	Title          string     `json:"title"`
	CompanyName    string     `json:"company_name"`
	AbsoluteURL    string     `json:"absolute_url"`
	InternalJobID  int        `json:"internal_job_id"`
	Location       location   `json:"location"`
	Metadata       []metadata `json:"metadata"`
	FirstPublished time.Time  `json:"first_published"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type location struct {
	Name string `json:"name"`
}

type metadata struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Value     metadataValue `json:"value"`
	ValueType string        `json:"value_type"`
}

type metadataValue struct {
	String    string
	StringArr []string
}

func (mv *metadataValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var stringVal string
	if err := json.Unmarshal(data, &stringVal); err != nil {
		mv.String = stringVal
		return nil
	}

	var stringArr []string
	if err := json.Unmarshal(data, &stringArr); err != nil {
		mv.StringArr = stringArr
		return nil
	}
	return fmt.Errorf("unexpected value for metadata: %s", string(data))
}
