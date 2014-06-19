package aranGO

import (
	"reflect"
	"strings"
)

type Error map[string]string

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
		tag = structField.Tag.Get(key)
		if tag != "" {
			tags[structField.Name] = tag
		}
	}

	return tags
}
