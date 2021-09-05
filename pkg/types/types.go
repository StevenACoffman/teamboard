package types

import (
	"encoding/json"
	"time"
)

type Repository struct {
	NameWithOwner string `json:"nameWithOwner"`
}
type Node struct {
	Typename     string      `json:"__typename"`
	Number       int         `json:"number"`
	Title        string      `json:"title"`
	Repository   Repository  `json:"repository"`
	CreatedAt    time.Time   `json:"createdAt"`
	MergedAt     interface{} `json:"mergedAt"`
	URL          string      `json:"url"`
	ChangedFiles int         `json:"changedFiles"`
	Additions    int         `json:"additions"`
	Deletions    int         `json:"deletions"`
}
type Edges struct {
	Node Node `json:"node"`
}

func (n Node) UnmarshalJSON(data []byte) error {
	var v Node
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}


	n.Typename=     v.Typename
	n.Number=       v.Number
	n.Title=        v.Title
	n.Repository=   v.Repository
	n.CreatedAt=    v.CreatedAt
	n.MergedAt=     v.MergedAt
	n.URL=          v.URL
	n.ChangedFiles= v.ChangedFiles
	n.Additions=    v.Additions
	n.Deletions=    v.Deletions


	return nil
}
func (n Node) MarshalJSON() ([]byte, error) {
	return json.Marshal(n)
}