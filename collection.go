package aranGO

// TODO Must Implement revision control

import (
	"errors"
	nap "github.com/jmcvetta/napping"
	"reflect"
)

type Collection struct {
	db     *Database `json:"db"`
	Name   string    `json:"name"`
	System bool      `json:"isSystem"`
	Status int       `json:"status"`
	// 3 = Edges , 2 =  Documents
	Type int `json:"type"`

	// last ,
	policy   string `json:"-"`
	revision bool   `json:"-"`
}

// Save saves doc into collection, doc should have Document Embedded to retrieve error and Key later.
func (col *Collection) Save(doc interface{}) error {
	var err error
	var res *nap.Response

	if col.Type == 2 {
		res, err = col.db.send("document?collection="+col.Name, "", "POST", doc, &doc, &doc)
	} else {
		return errors.New("Trying to save doc into EdgeCollection")
	}

	if err != nil {
		return err
	}

	if res.Status() != 201 && res.Status() != 202 {
		return errors.New("Unable to save document error")
	}

	return nil
}

// Save Edge into Edges collection
func (col *Collection) SaveEdge(doc interface{}, from string, to string) error {
	var err error
	var res *nap.Response

	if col.Type == 3 {
		res, err = col.db.send("edge?collection="+col.Name+"&from="+from+"&to="+to, "", "POST", doc, &doc, &doc)
	} else {
		return errors.New("Trying to save document into Edge-Collection")
	}

	if err != nil {
		return err
	}

	if res.Status() != 201 && res.Status() != 202 {
		return errors.New("Unable to save document error")
	}

	return nil

}

// Relate documents in edge collection
func (col *Collection) Relate(from *Document, to *Document, label interface{}) error {
	if col.Type == 2 {
		return errors.New("Invalid collection to add Edge: " + col.Name)
	}
	if from.Id == "" || to.Id == "" {
		return errors.New("from or to documents don't exist")
	}

	if from == nil || to == nil {
		return errors.New("Invalid document to link")
	}

	return col.SaveEdge(label, from.Id, to.Id)

}

//Get Document
func (col *Collection) Get(key string, doc interface{}) error {
	var err error

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		_, err = col.db.get("document", col.Name+"/"+key, "GET", nil, &doc, &doc)
	} else {
		_, err = col.db.get("edge", col.Name+"/"+key, "GET", nil, &doc, &doc)
	}

	if err != nil {
		return err
	}

	return nil
}

// Replace document
func (col *Collection) Replace(key string, doc interface{}) error {
	var err error
	var res *nap.Response

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		res, err = col.db.send("document", col.Name+"/"+key, "PUT", doc, &doc, &doc)
	} else {
		res, err = col.db.send("edge", col.Name+"/"+key, "PUT", doc, &doc, &doc)
	}

	if err != nil {
		return err
	}

	if res.Status() != 201 {
		return errors.New("Unable to replace document")
	}

	return nil
}

func (col *Collection) Patch(key string, doc interface{}) error {
	var err error
	var res *nap.Response

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		res, err = col.db.send("document", col.Name+"/"+key, "PATCH", doc, &doc, &doc)
	} else {
		res, err = col.db.send("edge", col.Name+"/"+key, "PATCH", doc, &doc, &doc)
	}

	if err != nil {
		return err
	}

	if res.Status() != 201 {
		return errors.New("Unable to replace document")
	}

	return nil
}

func (col *Collection) Delete(key string) error {
	var err error
	var res *nap.Response

	if key == "" {
		return errors.New("Key must not be empty")
	}

	if col.Type == 2 {
		res, err = col.db.get("document", col.Name+"/"+key, "DELETE", nil, nil, nil)
	} else {
		res, err = col.db.get("edge", col.Name+"/"+key, "DELETE", nil, nil, nil)
	}
	if err != nil {
		return err
	}

	switch res.Status() {
	case 202, 200:
		return nil
	default:
		return errors.New("Document don't exist or revision error")

	}
}

func (col *Collection) Exist() bool {
	return true
}

func Collections(db *Database) error {
	var err error
	var res *nap.Response

	// get all non-system collections
	res, err = db.get("collection?excludeSystem=true", "", "GET", nil, db, db)
	if err != nil {
		return err
	}

	if res.Status() == 200 {
		return nil
	} else {
		return errors.New("Failed to retrieve collections from Database")
	}
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}
	return val
}

// check if a key is unique or not
func (c *Collection) Unique(key string, value interface{}, index string) (bool, error) {
	var cur *Cursor
	var err error
	switch index {
	case "hash":
		// must implement other simple query function s
	case "skip-list":

	default:
		cur, err = c.Example(map[string]interface{}{key: value}, 0, 2)
	}
	if err != nil {
		return false, err
	}

	if cur.Amount > 0 {
		return false, nil
	}

	return true, nil
}

// Simple Queries

func (c *Collection) All(skip, limit int) (*Cursor, error) {
	var cur Cursor
	query := map[string]interface{}{"collection": c.Name, "skip": skip, "limit": limit}
	// sernd request
	res, err := c.db.send("simple", "all", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

func (c *Collection) Example(doc interface{}, skip, limit int) (*Cursor, error) {
	var cur Cursor
	query := map[string]interface{}{"collection": c.Name, "example": doc, "skip": skip, "limit": limit}
	// sernd request
	res, err := c.db.send("simple", "by-example", "PUT", query, &cur, &cur)

	if err != nil {
		return nil, err
	}

	if res.Status() == 201 {
		return &cur, nil
	} else {
		return nil, errors.New("Failed to execute query")
	}
}

func (c *Collection) First(example, doc interface{}) error {
	query := map[string]interface{}{"collection": c.Name, "example": doc}

	// sernd request
	res, err := c.db.send("simple", "first-example", "PUT", query, &doc, &doc)

	if err != nil {
		return err
	}

	if res.Status() == 200 {
		return nil
	} else {
		return errors.New("Failed to execute query")
	}

}
