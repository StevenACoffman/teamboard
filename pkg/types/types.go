package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Repository includes the requested fields of the GraphQL type Repository.
// The GraphQL type's documentation follows.
//
// A repository contains the content for a project.
type Repository struct {
	NameWithOwner string `json:"nameWithOwner"`
}

type Item interface {
	implementsItem()
	// GetTypename returns the receiver's concrete GraphQL type-name (see interface doc for possible values).
	GetTypename() string
}

func (v *EdgeNodePullRequest) implementsItem() {
}

// GetTypename is a part of, and documented with, the interface Item.
func (v *EdgeNodePullRequest) GetTypename() string {
	return v.Typename
}

// EdgeNodePullRequest includes the requested fields of the GraphQL type PullRequest.
// The GraphQL type's documentation follows.
//
// A repository pull request.
type EdgeNodePullRequest struct {
	Typename string `json:"__typename"`
	// Identifies the pull request number.
	Number int `json:"number"`
	// Identifies the pull request title.
	Title string `json:"title"`
	// The repository associated with this node.
	Repository Repository `json:"repository"`
	// Identifies the date and time when the object was created.
	CreatedAt time.Time `json:"createdAt"`
	// The date and time that the pull request was merged.
	MergedAt time.Time `json:"mergedAt"`
	// The HTTP URL for this pull request.
	Url string `json:"url"`
	// The number of changed files in this pull request.
	ChangedFiles int `json:"changedFiles"`
	// The number of additions in this pull request.
	Additions int `json:"additions"`
	// The number of deletions in this pull request.
	Deletions int `json:"deletions"`
}

func __unmarshalEdgeNodeSearchResultItem(v *Item, m json.RawMessage) error {
	if string(m) == "null" {
		return nil
	}

	var tn struct {
		TypeName string `json:"__typename"`
	}
	err := json.Unmarshal(m, &tn)
	if err != nil {
		return err
	}

	switch tn.TypeName {

	case "PullRequest":
		*v = new(EdgeNodePullRequest)
		return json.Unmarshal(m, *v)

	case "":
		return fmt.Errorf(
			"Response was missing SearchResultItem.__typename")
	default:
		return fmt.Errorf(
			`Unexpected concrete type for Item: "%v"`, tn.TypeName)
	}
}

func (v *Edge) UnmarshalJSON(b []byte) error {
	type EdgeWrapper Edge

	var firstPass struct {
		*EdgeWrapper
		Node json.RawMessage `json:"node"`
	}
	firstPass.EdgeWrapper = (*EdgeWrapper)(v)

	err := json.Unmarshal(b, &firstPass)
	if err != nil {
		return err
	}

	{
		target := &v.Node
		raw := firstPass.Node
		err = __unmarshalEdgeNodeSearchResultItem(
			target, raw)
		if err != nil {
			return fmt.Errorf(
				"Unable to unmarshal Edge.Node: %w", err)
		}
	}
	return nil
}

// Connection includes the requested fields of the GraphQL type SearchResultItemConnection.
// The GraphQL type's documentation follows.
//
// A list of results that matched against a search query.
type Connection struct {
	// The number of issues that matched the search query.
	IssueCount int `json:"issueCount"`
	// A list of edges.
	Edges []Edge `json:"edges"`
}

// Edge includes the requested fields of the GraphQL type SearchResultItemEdge.
// The GraphQL type's documentation follows.
//
// An edge in a connection.
type Edge struct {
	// The item at the end of the edge.
	Node Item `json:"-"`
}
