package aranGO

import (
	"errors"
	"strings"
)

type Document struct {
	Id  string `json:"_id,omitempty"  `
	Rev string `json:"_rev,omitempty" `
	Key string `json:"_key,omitempty" `

	Error   bool   `json:"error,omitempty"`
	Message string `json:"errorMessage,omitempty"`
	Code    int    `json:"code,omitempty"`
	Num     int    `json:"errorNum,omitempty"`
}

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

// Check if document exist
func (d *Document) Exist(db *Database) (bool, error) {
	if db == nil {
		return false, errors.New("Invalid db")
	}
	return true, nil
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
