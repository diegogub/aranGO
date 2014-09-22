package aranGO

import (
	"errors"
	nap "github.com/diegogub/napping"
	"regexp"
	"time"
)

// Database struct
type Database struct {
	Name        string `json:"name"`
	Id          string `json:"id"`
	Path        string `json:"path"`
	System      bool   `json:"isSystem"`
	Collections []Collection
	sess        *Session
	baseURL     string
}

/*
type DatabaseResult struct {
	Result []string `json:"result"`
	Error  bool     `json:"error"`
	Code   int      `json:"code"`
}
*/

// Execute AQL query into server and returns cursor struct
func (d *Database) Execute(q *Query) (*Cursor, error) {
	if q == nil {
		return nil, errors.New("Cannot execute nil query")
	} else {
		// check if I need to validate query
		if q.Validate {
			if !d.IsValid(q) {
				return nil, errors.New(q.ErrorMsg)
			}
		}
		// create cursor
		c := NewCursor(d)
		t0 := time.Now()
		_, err := d.send("cursor", "", "POST", q, c, c)
		t1 := time.Now()
		if err != nil {
			return nil, err
		}
		c.max = len(c.Result) - 1
		c.Time = t1.Sub(t0)
		return c, nil
	}
}

// ExecuteTran executes transaction into the database
func (d *Database) ExecuteTran(t *Transaction) error {
	if t.Action == "" {
		return errors.New("Action must not be nil")
	}

	// record execution time
	t0 := time.Now()
	resp, err := d.send("transaction", "", "POST", t, t, t)
	if err != nil {
		panic(err)
	}
	t1 := time.Now()
	t.Time = t1.Sub(t0)

	if err != nil {
		return err
	}

	if resp.Status() == 400 {
		return errors.New("Error executing transaction")
	}

	return nil
}

func (d *Database) IsValid(q *Query) bool {
	if q != nil {
		res, err := d.send("query", "", "POST", map[string]string{"query": q.Aql}, q, q)
		if err != nil {
			return false
		}
		if res.Status() == 200 {
			return true
		} else {
			// could check error into query
			return false
		}
	} else {
		return false
	}
}

// Do a request to test if the database is up and user authorized to use it
func (d *Database) get(resource string, id string, method string, param *nap.Params, result, err interface{}) (*nap.Response, error) {
	url := d.buildRequest(resource, id)
	var r *nap.Response
	var e error

	switch method {
	case "OPTIONS":
		r, e = d.sess.nap.Options(url, result, err)
	case "HEAD":
		r, e = d.sess.nap.Head(url, result, err)
	case "DELETE":
		r, e = d.sess.nap.Delete(url, result, err)
	default:
		r, e = d.sess.nap.Get(url, param, result, err)
	}

	return r, e
}

func (d *Database) send(resource string, id string, method string, payload, result, err interface{}) (*nap.Response, error) {
	url := d.buildRequest(resource, id)
	var r *nap.Response
	var e error

	switch method {
	case "POST":
		r, e = d.sess.nap.Post(url, payload, result, err)
	case "PUT":
		r, e = d.sess.nap.Put(url, payload, result, err)
	case "PATCH":
		r, e = d.sess.nap.Patch(url, payload, result, err)
	}
	return r, e
}

func (db Database) buildRequest(t string, id string) string {
	var r string
	if id == "" {
		r = db.baseURL + t
	} else {
		r = db.baseURL + t + "/" + id
	}
	return r
}

// Col returns Collection attached to current Database
func (db Database) Col(name string) *Collection {
	var col Collection
	var found bool
	// need to validate this more
	for _, c := range db.Collections {
		if c.Name == name {
			col = c
			col.db = &db
			found = true
			break
		}
	}

	if !found {
		if db.sess.safe {
			panic("Collection " + name + " not found")
		} else {
			var col CollectionOptions
			col.Name = name
			db.CreateCollection(&col)
			return db.Col(name)
		}
	}
	return &col
}

func validColName(name string) error {
	reg, err := regexp.Compile(`^[A-z]+[0-9\-_]*`)

	if err != nil {
		return err
	}
	if !reg.MatchString(name) {
		return errors.New("Invalid collection name")
	}

	return nil
}

// Create collections
func (d *Database) CreateCollection(c *CollectionOptions) error {

	err := validColName(c.Name)
	if err != nil {
		return err
	}

	resp, err := d.send("collection", "", "POST", c, nil, nil)
	if err != nil {
		return err
	}

	switch resp.Status() {
	case 200:
		//push name into list
		Collections(d)
		return nil
	default:
		return errors.New("Failed to create collection")
	}
}

//Drop Collection
func (d *Database) DropCollection(name string) error {
	resp, err := d.get("collection", name, "DELETE", nil, nil, nil)

	if err != nil {
		return err
	}

	switch resp.Status() {
	case 200:
		return nil
	default:
		return errors.New("Failed to create collection")
	}
}

// Truncate collection
func (d *Database) TruncateCollection(name string) error {
	resp, err := d.send("collection", name+"/truncate", "PUT", nil, nil, nil)

	if err != nil {
		return err
	}
	switch resp.Status() {
	// TODO need to define return codes
	case 201:
		return nil
	case 200:
		return nil
	case 202:
		return nil
	default:
		return errors.New("Failed to truncate collection")
	}
}

// ColExist checks if collection exist
func (db *Database) ColExist(name string) bool {
	if name == "" {
		return false
	}
	res, err := db.get("collection", name, "GET", nil, nil, nil)
	if err != nil {
		panic(err)
	}

	switch res.Status() {
	case 404:
		return false
	default:
		return true
	}
}

// CheckCollection returns collection option based on name, nil otherwise
func (d *Database) CheckCollection(name string) *CollectionOptions {
	var col CollectionOptions
	if name == "" {
		return nil
	}

	resp, err := d.get("collection", name, "GET", nil, &col, &col)
	if err != nil {
		return nil
	}

	if resp.Status() == 200 {
		return &col
	}
	return nil
}
