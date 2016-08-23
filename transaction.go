package aranGO

import (
	"errors"
	"io/ioutil"
	"time"
)

type Transaction struct {
	Collections map[string][]string `json:"collections"`
	Action      string              `json:"action"`
	Result      interface{}         `json:"result,omitempty"`

	//Optional
	Sync      bool                   `json:"waitForSync,omitempty"`
	Lock      int                    `json:"lockTimeout,omitempty"`
	Replicate bool                   `json:"replicate,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Time      time.Duration          `json:"time,omitempty"`

	//ErrorInfo
	Error bool `json:"error,omitempty"`
	Code  int  `json:"code,omitempty"`
	Num   int  `json:"errorNum,omitempty"`
}

func NewTransaction(q string, write []string, read []string) *Transaction {
	var t Transaction
	t.Collections = make(map[string][]string)
	t.Action = q
	if write != nil {
		t.Collections["write"] = write
	}
	if read != nil {
		t.Collections["read"] = read
	}
	return &t
}

func (t *Transaction) FromFile(path string) error {
	var err error
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	t.Action = string(b)

	return err
}

func (t *Transaction) Execute(db *Database) error {
	if db == nil {
		return errors.New("Nil database")
	}
	err := db.ExecuteTran(t)
	return err
}
