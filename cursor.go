package aranGO

import (
	"encoding/json"
	"errors"
	"reflect"
)

type Cursor struct {
	db *Database `json:"-"`
	Id string    `json:"Id"`

	Index  int           `json:"-"`
	Result []interface{} `json:"result"`
	More   bool          `json:"hasMore"`
	Amount int           `json:"count"`
	Data   Extra         `json:"extra"`

	Err    bool   `json:"error"`
	ErrMsg string `json:"errorMessage"`
	Code   int    `json:"code"`
}

func NewCursor(db *Database) *Cursor {
	var c Cursor
	if db == nil {
		return nil
	}
	c.db = db
	return &c
}

func (c *Cursor) FetchBatch(r interface{}) error {
	kind := reflect.ValueOf(r).Elem().Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return errors.New("Container must be Slice of array kind")
	}
	b, err := json.Marshal(c.Result)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, r)
	if err != nil {
		return err
	}

	return nil
}

// Iterates over cursor, returns false when no more values into row, fetch next batch if necesary.
func (c *Cursor) FetchOne(r interface{}) (bool, error) {
	var max int = len(c.Result) - 1

	if c.Index < max {
		b, err := json.Marshal(c.Result[c.Index])
		err = json.Unmarshal(b, r)
		c.Index++ // move to next value into result
		if err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
	// Last document
	if c.Index == max {
		b, err := json.Marshal(c.Result[c.Index])
		err = json.Unmarshal(b, r)
		if err != nil {
			return false, err
		} else {
			if c.More {
				//fetch rest from server
				res, err := c.db.send("cursor", c.Id, "PUT", nil, c, c)

				if err != nil {
					return false, err
				}

				if res.Status() == 200 {
					c.Index = 0
					return true, nil
				} else {
					return false, nil
				}

			} else {
				// last doc
				return false, nil
			}
		}
	}

	return false, nil
}

type Extra struct {
	FullCount int `json:"fullCount"`
}

func (c Cursor) Count() int {
	return c.Amount
}

func (c *Cursor) FullCount() int {
	return c.Data.FullCount
}

func (c Cursor) HasMore() bool {
	return c.More
}

func (c Cursor) Error() bool {
	return c.Err
}

func (c Cursor) ErrCode() int {
	return c.Code
}
