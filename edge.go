package aranGO

import "errors"

type Edge struct {
	Id		string	`json:"_id,omitempty"`
	Key		string	`json:"_key,omitempty"`
	Rev		string	`json:"_rev,omitempty"`
	Oldrev	string	`json:"_oldRev,omitempty"`
	From	string	`json:"_from"`
	To		string	`json:"_to"`
	Label	string	`json:"$label,omitempty"` // Not realized in v3.2

	Error   bool	`json:"error,omitempty"`
	Message string	`json:"errorMessage,omitempty"`
}

type edgeStats struct {
	Scanned		int	`json:"scannedIndex"`
	Filtered	int	`json:"filtered"`
}

type edgesResponce struct {
	Edges	[]Edge		`json:"edges"`
	Stats	edgeStats	`json:"stats"`
	
	Code	int			`json:"code"`
	Error   bool		`json:"error"`

}
// Read in- or outbound edges
// GET /_api/edges/{collection-id}
func (db *Database) Edges(edgeCol string, startVertex string, direction string) ([]Edge, error) {
	if edgeCol == "" || startVertex == "" {
		return nil, errors.New("Invalid key or collection")
	}

	// TODO: Make URL parce in DB communication functions
	uri := edgeCol + "?vertex=" + startVertex
	if direction == "in" || direction == "out" {
		uri += "&direction=" + direction
	}

	var eresp edgesResponce
	res, err := db.get("edges", uri, "GET", nil, &eresp, &eresp)
	if err != nil {
		return nil, err
	}

	switch res.Status() {
	case 200:
		return eresp.Edges, nil
	case 400:
		return nil, errors.New("Request contains invalid parameters")
	case 404:
		return nil, errors.New("Edge collection was not found")
	default:
		return nil, errors.New("DB error")
	}
}