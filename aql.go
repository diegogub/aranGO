package aranGO

import (
    "errors"
    "strconv"
)

// Aql query
type Query struct {
	// mandatory
	Aql string `json:"query,omitempty"`
	//Optional values
	Batch    int                    `json:"batchSize,omitempty"`
	Count    bool                   `json:"count,omitempty"`
	BindVars map[string]interface{} `json:"bindVars,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
  // opetions fullCount bool
  // Note that the fullCount sub-attribute will only be present in the result if the query has a LIMIT clause and the LIMIT clause is actually used in the query.
	// Control
	Validate bool   `json:"-"`
	ErrorMsg string `json:"errorMessage,omitempty"`
}

func NewQuery(query string) *Query {
  var q Query
  // alocate maps
  q.Options = make(map[string]interface{})
  q.BindVars= make(map[string]interface{})

	if query == "" {
		return &q
	} else {
		q.Aql = query
		return &q
	}
}

func (q *Query) Modify(query string) error {
	if query == "" {
		return errors.New("query must not be empty")
	} else {
		q.Aql = query
		return nil
	}
}
