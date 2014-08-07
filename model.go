package aranGO

import (
	"reflect"
	"strings"
)

type Error map[string]string

func NewError() Error {
  var err Error
  err = make(map[string]string)
  return err
}

type Modeler interface {
	// Returns current model key
	GetKey() string
	// Returns collection where I should save the model
	GetCollection() string
	// Error
	GetError() (string, bool)
	// hooks
	PreSave(err Error)
	PostSave(err Error)
	PreUpdate(err Error)
	PostUpdate(err Error)
	PreDelete(err Error)
	PostDelete(err Error)
}

// Updates or save new Model
func Save(db *Database, m Modeler) Error {
	var err Error = make(map[string]string)
	col := m.GetCollection()
	key := m.GetKey()

	// basic validation
	validate(m, db, col, err)
	if len(err) > 0 {
		return err
	}

	if key == "" {
		m.PreSave(err)
		if len(err) > 0 {
			return err
		}
		e := db.Col(col).Save(m)
		if e != nil {
			// db error
			err["db"] = e.Error()
		}
		// check if model has errors
		docerror, haserror := m.GetError()
		if haserror {
			err["doc"] = docerror
			return err
		}

		m.PostSave(err)
	} else {
		m.PreUpdate(err)
		if len(err) > 0 {
			return err
		}

		e := db.Col(col).Replace(key, m)
		if e != nil {
			// db error
			err["db"] = e.Error()
		}

		// check if model has errors
		docerror, haserror := m.GetError()
		if haserror {
			err["doc"] = docerror
			return err
		}
		m.PostUpdate(err)
	}

	return err
}

func Delete(db *Database, m Modeler) Error {
	var err Error
	key := m.GetKey()
	col := m.GetCollection()
	if key == "" {
		//
		err["key"] = "invalid"
		return err
	}
	// pre delete hook
	m.PreDelete(err)
	if len(err) > 0 {
		return err
	}
	e := db.Col(col).Delete(key)
	if e != nil {
		err["db"] = e.Error()
	}
	docerror, haserror := m.GetError()
	if haserror {
		err["doc"] = docerror
		return err
	}

	m.PostDelete(err)

	return err
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
				// check if it's empty, depending on Kind
				switch field.Kind() {
				case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
					if field.Len() == 0 {
						err[fname] = "required"
					}
				case reflect.Interface, reflect.Ptr:
					if field.IsNil() {
						err[fname] = "required"
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
			f := field.FieldByName(fname)
			valid = false
			for _, e := range enumValues {
				if e == f.String() {
					valid = true
				}
			}
			if !valid {
				err[fname] = "invalid"
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
