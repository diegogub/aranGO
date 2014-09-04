package aranGO

import (
	"reflect"
  "errors"
	"strings"
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
  keys map[string]interface{}
  db *Database
  err Error
}

func NewContext(db *Database) (*Context,error){
  if db == nil  {
    return nil,errors.New("Invalid DB")
  }
  var c Context
  c.db = db
  c.keys = make(map[string]interface{})
  c.err = make(map[string]string)

  return &c,nil
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
type PreSaver interface{
	PreSave(c *Context)
}

type PostSaver interface{
	PostSave(c *Context)
}

type PreUpdater interface{
	PreUpdate(c *Context)
}

type PostUpdater interface{
	PostUpdate(c *Context)
}

type PreDeleter interface{
	PreDelete(c *Context)
}

type PostDeleter interface{
	PostDelete(c *Context)
}

// Updates or save new Model
func (c *Context) Save(m Modeler) Error {
	col := m.GetCollection()
	key := m.GetKey()

	// basic validation
	validate(m, c.db, col, c.err)
	if len(c.err) > 0 {
		return c.err
	}

	if key == "" {

    if hook, ok := m.(PreSaver); ok{
		  hook.PreSave(c)
    }

		if len(c.err) > 0 {
			return c.err
		}
    setTimes(m.(interface{}),"save")
		e := c.db.Col(col).Save(m)
		if e != nil {
			// db c.error
			c.err["db"] = e.Error()
		}
		// check if model has errors
		docerror, haserror := m.GetError()
		if haserror {
			c.err["error"] = docerror
			return c.err
		}

    if hook, ok := m.(PostSaver); ok{
		  hook.PostSave(c)
    }

	} else {

    if hook, ok := m.(PreUpdater); ok{
		  hook.PreUpdate(c)
    }

		if len(c.err) > 0 {
			return c.err
		}

		e := c.db.Col(col).Replace(key, m)
		if e != nil {
			// db error
			c.err["db"] = e.Error()
		}

		// check if model has errors
		docerror, haserror := m.GetError()
		if haserror {
			c.err["doc"] = docerror
			return c.err
		}

    if hook, ok := m.(PostUpdater); ok{
		  hook.PostUpdate(c)
    }
	}

	return c.err
}

func (c *Context) Delete(db *Database, m Modeler) Error {
	key := m.GetKey()
	col := m.GetCollection()
	if key == "" {
		//
		c.err["key"] = "invalid"
		return c.err
	}
	// pre delete hook
  if hook, ok := m.(PreDeleter); ok{
    hook.PreDelete(c)
  }
	if len(c.err) > 0 {
		return c.err
	}
	e := db.Col(col).Delete(key)
	if e != nil {
		c.err["db"] = e.Error()
	}
	docerror, haserror := m.GetError()
	if haserror {
		c.err["doc"] = docerror
		return c.err
	}

  if hook, ok := m.(PostDeleter); ok{
    hook.PostDelete(c)
  }

	return c.err
}

func Unique(m interface{},db *Database,update bool, err Error){
  val := Tags(m,"unique")
  var uniq bool
  fvalue := reflectValue(m)
  for fname, col:= range val {
    field := fvalue.FieldByName(fname)
    fty := fvalue.Type()
    ftype ,ok:= fty.FieldByName(fname)
    if ok {
      if ftype.Anonymous && ftype.Type.Kind() == reflect.Struct {
        unique(field,val,db,&uniq,update,err)
      }else{
      // search by example
        c := db.Col(col)
        uniq , _ = c.Unique(fname,field.String(),update,"")
      }
      if !uniq{
        err[fname] = "not unique"
      }
    }
  }
}

func unique(m reflect.Value,val map[string]string,db *Database,uniq *bool,update bool,err Error){
  for fname, col:= range val {
    field := reflectValue(m).FieldByName(fname)
    ftype ,ok:= field.Type().FieldByName(fname)
    if ok {
      if ftype.Anonymous && ftype.Type.Kind() == reflect.Struct {
        unique(field,val,db,uniq,update,err)
      }else{
      // search by example
        *uniq, _ = db.Col(col).Unique(fname,field.String(),update,"")
      }
      if !*uniq{
        err[fname] = "not unique"
      }
    }
  }
}

func Validate(m interface{}, db *Database,col string, err Error){
	checkRequired(m, err)
	checkEnum(m, err)

	val := Tags(m, "sub")
	if len(val) > 0 {
		for fname, _ := range val {
			field := reflectValue(m).FieldByName(fname)
			// All sub structures are not Models
			validate(field.Interface(), db, col, err)
		}
	}
	return
}

func validate(m interface{}, db *Database, col string, err Error) {
	checkRequired(m, err)
	checkEnum(m, err)

	val := Tags(m, "sub")
	if len(val) > 0 {
		for fname, _ := range val {
			field := reflectValue(m).FieldByName(fname)
			// All sub structures are not Models
			validate(field.Interface(), db, col, err)
		}
	}
	return
}

func checkUnique(m interface{},db *Database,update bool,err Error){
}

func checkRequired(m interface{}, err Error) {
	req := Tags(m, "required")
	if len(req) > 0 {
		for fname, _ := range req {
			if !checkField(m, fname) { // if don't
				err[fname] = "required"
			} else {
				field := reflectValue(m).FieldByName(fname)
        jname  := Tag(m,fname,"json")
				// check if it's empty, depending on Kind
				switch field.Kind() {
				case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
					if field.Len() == 0 {
            if jname == "" {
              err[fname] = "invalid"
            }else{
              err[jname] = "invalid"
            }
					}
				case reflect.Interface, reflect.Ptr:
					if field.IsNil() {
            if jname == "" {
              err[fname] = "invalid"
            }else{
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
      jname  := Tag(m,fname,"json")

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
        }else{
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
	objValue ,_:= reflect.TypeOf(obj).Elem().FieldByName(fname)
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
      getTags(structField.Type,tags,key)
    }else{
      tag = structField.Tag.Get(key)
      if tag != "" {
        tags[structField.Name] = tag
      }
    }
	}
	return tags
}

func getTags(obj reflect.Type,tags map[string]string,key string){

	if obj.Kind() != reflect.Struct && obj.Kind() != reflect.Ptr {
		return
	}

	var tag string

	fieldsCount := obj.NumField()

	for i := 0; i < fieldsCount; i++ {
		structField := obj.Field(i)
    if structField.Anonymous && structField.Type.Kind() == reflect.Struct {
      getTags(obj,tags,key)
    }else{
      tag = structField.Tag.Get(key)
      if tag != "" {
        tags[structField.Name] = tag
      }
    }
	}
}

func setTimes(obj interface{},action string){
  timeFields := Tags(obj,"time")
  if len(timeFields) > 0 {
    for fname , val := range timeFields {
      if val == action {
        f := reflectValue(obj).FieldByName(fname)
			  switch f.Kind(){
          case reflect.Int64:
            t:= time.Now().Unix() * 1000
            f.Set(reflect.ValueOf(t))
           default:
            t := time.Now().UTC()
            f.Set(reflect.ValueOf(t))
        }
      }
    }
  }
}
