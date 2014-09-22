package aranGO

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Error map[string]string

func NewError() Error {
	var err Error
	err = make(map[string]string)
	return err
}

// Context to share state between hook and track transaction state
type Context struct {
	Keys map[string]interface{}
	Db   *Database
	Err  Error
}

func NewContext(db *Database) (*Context, error) {
	if db == nil {
		return nil, errors.New("Invalid DB")
	}
	var c Context
	c.Db = db
	c.Keys = make(map[string]interface{})
	c.Err = make(map[string]string)

	return &c, nil
}

type Modeler interface {
	// Returns current model key
	GetKey() string
	// Returns collection where I should save the model
	GetCollection() string
	// Error
	GetError() (string, bool)
	// hooks
}

// hook interfaces
type PreSaver interface {
	PreSave(c *Context)
}

type PostSaver interface {
	PostSave(c *Context)
}

type PreUpdater interface {
	PreUpdate(c *Context)
}

type PostUpdater interface {
	PostUpdate(c *Context)
}

type PreDeleter interface {
	PreDelete(c *Context)
}

type PostDeleter interface {
	PostDelete(c *Context)
}

//Get model
func (c *Context) Get(m Modeler) Error {
	col := m.GetCollection()
	key := m.GetKey()

	c.Db.Col(col).Get(key, m)
	docerror, haserror := m.GetError()
	if haserror {
		c.Err["error"] = docerror
		return c.Err
	}

	return c.Err
}

// Updates or save new Model into database
func (c *Context) Save(m Modeler) Error {
	col := m.GetCollection()
	key := m.GetKey()

	// basic validation

	if key == "" {

		validate(m, c.Db, col, false, c.Err)
		if len(c.Err) > 0 {
			return c.Err
		}

		if hook, ok := m.(PreSaver); ok {
			hook.PreSave(c)
		}
		if len(c.Err) > 0 {
			return c.Err
		}

		setTimes(m.(interface{}), "save")
		e := c.Db.Col(col).Save(m)
		if e != nil {
			// db c.error
			c.Err["Db"] = e.Error()
		}
		// check if model has errors
		docerror, haserror := m.GetError()
		if haserror {
			c.Err["error"] = docerror
			return c.Err
		}

		if hook, ok := m.(PostSaver); ok {
			hook.PostSave(c)
		}

	} else {

		validate(m, c.Db, col, true, c.Err)
		if len(c.Err) > 0 {
			return c.Err
		}

		if hook, ok := m.(PreUpdater); ok {
			hook.PreUpdate(c)
		}

		if len(c.Err) > 0 {
			return c.Err
		}

		setTimes(m.(interface{}), "save")
		e := c.Db.Col(col).Replace(key, m)
		if e != nil {
			// db error
			c.Err["db"] = e.Error()
		}

		// check if model has errors
		docerror, haserror := m.GetError()
		if haserror {
			c.Err["doc"] = docerror
			return c.Err
		}

		if hook, ok := m.(PostUpdater); ok {
			hook.PostUpdate(c)
		}
	}

	return c.Err
}

type auxModelPos struct {
	pos int
	err Error
}

//Saves models into database concurrently
func (c *Context) BulkSave(models []Modeler) map[int]Error {
	var wg sync.WaitGroup
	errorMap := make(map[int]Error)
	ch := make(chan auxModelPos)
	ok := make(chan bool)

	wg.Add(len(models))
	for i, mod := range models {
		go func(i int, mod Modeler) {
			err := c.Save(mod)
			if len(err) > 0 {
				errPos := auxModelPos{pos: i, err: err}
				ch <- errPos
			} else {
				ok <- true
			}
		}(i, mod)
	}

	go func() {
		for {
			select {
			case p := <-ch:
				errorMap[p.pos] = p.err
				wg.Done()
			case <-ok:
				wg.Done()
			}
		}
	}()

	wg.Wait()
	return errorMap
}

func (c *Context) Delete(m Modeler) Error {
	key := m.GetKey()
	col := m.GetCollection()
	if key == "" {
		//
		c.Err["key"] = "invalid"
		return c.Err
	}
	// pre delete hook
	if hook, ok := m.(PreDeleter); ok {
		hook.PreDelete(c)
	}
	if len(c.Err) > 0 {
		return c.Err
	}
	e := c.Db.Col(col).Delete(key)
	if e != nil {
		c.Err["db"] = e.Error()
	}
	docerror, haserror := m.GetError()
	if haserror {
		c.Err["doc"] = docerror
		return c.Err
	}

	if hook, ok := m.(PostDeleter); ok {
		hook.PostDelete(c)
	}

	return c.Err
}

func Unique(m interface{}, db *Database, update bool, err Error) {
	val := Tags(m, "unique")
	var uniq bool
	fvalue := reflectValue(m)
	for fname, col := range val {
		field := fvalue.FieldByName(fname)
		fty := fvalue.Type()
		ftype, ok := fty.FieldByName(fname)
		if ok {
			if ftype.Anonymous && ftype.Type.Kind() == reflect.Struct {
				unique(field, val, db, &uniq, update, err)
			} else {
				// validate collection name!!!!
				validName := validColName(col)
				if col == "-" || col == "" || validName != nil {
					err["colname"] = "Invalid collection name in unique tag"
					return
				}
				c := db.Col(col)
				jname := Tag(m, fname, "json")
				if jname != "" {
					uniq, _ = c.Unique(jname, field.String(), update, "")
				} else {
					uniq, _ = c.Unique(fname, field.String(), update, "")
				}
			}
			if !uniq {
				err[fname] = "not unique"
			}
		}
	}
}

func unique(m reflect.Value, val map[string]string, db *Database, uniq *bool, update bool, err Error) {
	for fname, col := range val {
		field := reflectValue(m).FieldByName(fname)
		ftype, ok := field.Type().FieldByName(fname)
		if ok {
			if ftype.Anonymous && ftype.Type.Kind() == reflect.Struct {
				unique(field, val, db, uniq, update, err)
			} else {
				// search by example
				jname := Tag(m, fname, "json")
				if jname != "" {
					*uniq, _ = db.Col(col).Unique(jname, field.String(), update, "")
				} else {
					*uniq, _ = db.Col(col).Unique(fname, field.String(), update, "")
				}
			}
			if !*uniq {
				err[fname] = "not unique"
			}
		}
	}
}

func Validate(m interface{}, db *Database, col string, update bool, err Error) {
	checkRequired(m, err)
	checkEnum(m, err)
	Unique(m, db, update, err)

	val := Tags(m, "sub")
	if len(val) > 0 {
		for fname, _ := range val {
			field := reflectValue(m).FieldByName(fname)
			// All sub structures are not Models
			validate(field.Interface(), db, col, update, err)
		}
	}
	return
}

func validate(m interface{}, db *Database, col string, update bool, err Error) {
	checkRequired(m, err)
	checkEnum(m, err)
	Unique(m, db, update, err)

	val := Tags(m, "sub")
	if len(val) > 0 {
		for fname, _ := range val {
			field := reflectValue(m).FieldByName(fname)
			// All sub structures are not Models
			validate(field.Interface(), db, col, update, err)
		}
	}
	return
}

func checkUnique(m interface{}, db *Database, update bool, err Error) {
}

func checkRequired(m interface{}, err Error) {
	req := Tags(m, "required")
	if len(req) > 0 {
		for fname, _ := range req {
			if !checkField(m, fname) { // if don't
				err[fname] = "required"
			} else {
				field := reflectValue(m).FieldByName(fname)
				jname := Tag(m, fname, "json")
				// check if it's empty, depending on Kind
				switch field.Kind() {
				case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
					if field.Len() == 0 {
						if jname == "" {
							err[fname] = "invalid"
						} else {
							err[jname] = "invalid"
						}
					}
				case reflect.Interface, reflect.Ptr:
					if field.IsNil() {
						if jname == "" {
							err[fname] = "invalid"
						} else {
							err[jname] = "invalid"
						}
					}
				}
			}
		}
	}
	return
}

func checkEnum(m interface{}, err Error) {
	enumFields := Tags(m, "enum")
	if len(enumFields) > 0 {
		field := reflectValue(m)
		valid := false
		for fname, enuml := range enumFields {
			enumValues := strings.Split(enuml, ",")
			jname := Tag(m, fname, "json")

			f := field.FieldByName(fname)
			valid = false
			for _, e := range enumValues {
				if e == f.String() {
					valid = true
				}
			}
			if !valid {
				if jname == "" {
					err[fname] = "invalid"
				} else {
					err[jname] = "invalid"
				}
			}
		}
	}
	return
}

func checkField(m interface{}, fname string) bool {
	field := reflectValue(m).FieldByName(fname)
	if field == reflect.ValueOf(nil) {
		return false
	}
	return true
}

func Tag(obj interface{}, fname, key string) string {
	if reflect.TypeOf(obj).Kind() != reflect.Struct && reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return ""
	}
	objValue, _ := reflect.TypeOf(obj).Elem().FieldByName(fname)
	return objValue.Tag.Get(key)
}

func Tags(obj interface{}, key string) map[string]string {

	if reflect.TypeOf(obj).Kind() != reflect.Struct && reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil
	}

	var tag string

	objValue := reflectValue(obj)
	objType := objValue.Type()
	fieldsCount := objType.NumField()
	tags := make(map[string]string)

	for i := 0; i < fieldsCount; i++ {
		structField := objType.Field(i)
		if structField.Anonymous && structField.Type.Kind() == reflect.Struct {
			getTags(structField.Type, tags, key)
		} else {
			tag = structField.Tag.Get(key)
			if tag != "" {
				tags[structField.Name] = tag
			}
		}
	}
	return tags
}

func getTags(obj reflect.Type, tags map[string]string, key string) {

	if obj.Kind() != reflect.Struct && obj.Kind() != reflect.Ptr {
		return
	}

	var tag string

	fieldsCount := obj.NumField()

	for i := 0; i < fieldsCount; i++ {
		structField := obj.Field(i)
		if structField.Anonymous && structField.Type.Kind() == reflect.Struct {
			getTags(obj, tags, key)
		} else {
			tag = structField.Tag.Get(key)
			if tag != "" {
				tags[structField.Name] = tag
			}
		}
	}
}

func setTimes(obj interface{}, action string) {
	timeFields := Tags(obj, "time")
	if len(timeFields) > 0 {
		for fname, val := range timeFields {
			if val == action {
				f := reflectValue(obj).FieldByName(fname)
				switch f.Kind() {
				case reflect.Int64:
					t := time.Now().Unix() * 1000
					f.Set(reflect.ValueOf(t))
				default:
					t := time.Now().UTC()
					f.Set(reflect.ValueOf(t))
				}
			}
		}
	}
}

func ObjT(m Modeler) ObjTran {
	var obt ObjTran
	obt.Collection = m.GetCollection()
	obt.Obj = m
	return obt
}

type Relation struct {
	Obj ObjTran `json:"obj"   `
	// Relate to map[edgeCol]obj
	EdgeCol string                 `json:"edgcol"`
	Label   map[string]interface{} `json:"label" `
	Rel     []ObjTran              `json:"rel"   `
	Error   bool                   `json:"error" `
	Update  bool                   `json:"update"`

	Db *Database
}

type ObjTran struct {
	Collection string      `json:"c"`
	Obj        interface{} `json:"o"`
}

func (a *Relation) Commit() error {
	col := []string{a.Obj.Collection}
	if a.EdgeCol != "" {
		col = append(col, a.EdgeCol)
	}

	q := `function(p){
        var db = require('internal').db;
        try{
          if ( p["act"]["obj"]["o"].hasOwnProperty("_key") && db[p["act"]["obj"]["c"]].exists(p["act"]["obj"]["o"]["_key"]) ) {
            if ( p["act"]["update"] ) {
              p["act"]["obj"]["o"] = db[p["act"]["obj"]["c"]].replace(p["act"]["obj"]["o"]["_id"],p["act"]["obj"]["o"])
              p["act"]["obj"]["o"] = db[p["act"]["obj"]["c"]].document(p["act"]["obj"]["o"]["_id"])
            }
          }else{
            if (p["act"]["obj"]["o"]["_key"] == null || p["act"]["obj"]["o"]["_key"] == ""){
              p["act"]["obj"]["o"] = db[p["act"]["obj"]["c"]].save(p["act"]["obj"]["o"])
              p["act"]["obj"]["o"] = db[p["act"]["obj"]["c"]].document(p["act"]["obj"]["o"]["_id"])
            }else{
              throw("invalid main object id")
            }
          }
        }catch(err){
          p["act"]["error"] = true
          p["act"]["msg"]   = err
        }

        mainId = p["act"]["obj"]["o"]["_id"]

        for (i= 0 ;i<p["act"]["rel"].length;i++){
            if ( p["act"]["rel"][i]["o"].hasOwnProperty("_key") && db[p["act"]["rel"][i]["c"]].exists(p["act"]["rel"][i]["o"]["_key"]) ) {
              if ( p["act"]["update"] ) {
                p["act"]["rel"][i]["o"] = db[p["act"]["rel"][i]["c"]].replace(p["act"]["rel"][i]["o"]["_id"],p["act"]["rel"][i]["o"])
                p["act"]["rel"][i]["o"] = db[p["act"]["rel"][i]["c"]].document(p["act"]["rel"][i]["o"]["_id"])
              }
            }else{
              if (p["act"]["rel"][i]["o"]["_key"] == null || p["act"]["rel"][i]["o"]["_key"] == ""){
                p["act"]["rel"][i]["o"] = db[p["act"]["rel"][i]["c"]].save(p["act"]["rel"][i]["o"])
                p["act"]["rel"][i]["o"] = db[p["act"]["rel"][i]["c"]].document(p["act"]["rel"][i]["o"]["_id"])
              }else{
                throw("invalid main relect id")
              }
            }
            // relate documents
            switch (p["act"]["dire"]){
              case "out":
                db[p["act"]["edgcol"]].save(mainId,p["act"]["rel"][i]["o"]["_id"],p["act"]["label"])
              case "in":
                db[p["act"]["edgcol"]].save(p["act"]["rel"][i]["o"]["_id"],mainId,p["act"]["label"])
              default:
                db[p["act"]["edgcol"]].save(mainId,p["act"]["rel"][i]["o"]["_id"],p["act"]["label"])
            }
        }

        return p["act"]
    }
  `

	trx := NewTransaction(q, col, nil)
	trx.Params = map[string]interface{}{"act": a}
	err := trx.Execute(a.Db)
	// Tedious unmarshaling. I should map, maps => struct
	b, _ := json.Marshal(trx.Result)
	json.Unmarshal(b, a)
	return err
}

func (c *Context) NewRelation(main Modeler, label map[string]interface{}, edgecol string, dierection string, rel ...Modeler) (*Relation, Error) {
	var act Relation
	key := main.GetKey()
	if key == "" {
		validate(main, c.Db, main.GetCollection(), false, c.Err)
	} else {
		validate(main, c.Db, main.GetCollection(), true, c.Err)
	}

	if len(c.Err) > 0 {
		return nil, c.Err
	}

	act.Obj = ObjT(main)
	act.Rel = make([]ObjTran, 0)
	act.Label = label
	act.EdgeCol = edgecol

	for _, mod := range rel {
		key := mod.GetKey()
		if key == "" {
			validate(mod, c.Db, mod.GetCollection(), false, c.Err)
		} else {
			validate(mod, c.Db, mod.GetCollection(), true, c.Err)
		}

		if len(c.Err) > 0 {
			return nil, c.Err
		}

		act.Rel = append(act.Rel, ObjT(mod))
	}
	act.Db = c.Db
	return &act, nil
}

// increase value by n
func Inc(field string, n int64) error {
	return nil
}
