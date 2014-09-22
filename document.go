package aranGO

import (
	"errors"
	"strings"
)

type Document struct {
	Id  string `json:"_id,omitempty"              `
	Rev string `json:"_rev,omitempty"             `
	Key string `json:"_key,omitempty"             `

	Error   bool   `json:"error,omitempty"`
	Message string `json:"errorMessage,omitempty"`
}

// Creates base document structure
func NewDocument(id string) (*Document, error) {
	// some basic validation
	sid := strings.Split(id, "/")
	if len(sid) != 2 {
		return nil, errors.New("Invalid id")
	}
	if id == "" {
		return nil, errors.New("Invalid empty id")
	}
	var d Document
	d.Id = id
	d.Key = sid[1]
	return &d, nil
}

// Return map[string]string of document instead of struct
func (d *Document) Map(db *Database) (map[string]string, error) {
	var m map[string]string
	sid := strings.Split(d.Id, "/")
	m = make(map[string]string)
	err := db.Col(sid[0]).Get(sid[1], &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func (d *Document) SetKey(key string) error {
	//valitated key
	d.Key = key
	return nil
}

func (d *Document) SetRev(rev string) error {
	d.Rev = rev
	return nil
}

// Check if a document was updated
func (d *Document) Updated(db *Database) (bool, error) {
	if db == nil {
		return false, errors.New("Invalid db")
	}
	// check document id and rev
	if d.Id == "" || d.Rev == "" {
		return false, errors.New("Document must exist or have valid _rev and _id")
	}
	// add revision id
	res, err := db.get("document", d.Id+"?rev="+d.Rev, "GET", nil, nil, nil)

	if err != nil {
		return false, err
	}

	switch res.Status() {
	case 404:
		return true, nil
	case 412:
		return true, nil
	default:
		return false, nil
	}
}

// Check if document exist
func (d *Document) Exist(db *Database) (bool, error) {

	if db == nil {
		return false, errors.New("Invalid db")
	}
	// check document id and rev
	if d.Id == "" {
		return false, errors.New("Document must exist or have valid _rev and _id")
	}
	// add revision id
	res, err := db.get("document", d.Id, "GET", nil, nil, nil)

	if err != nil {
		return false, err
	}

	switch res.Status() {
	case 404:
		return false, nil
	default:
		return true, nil
	}
}
