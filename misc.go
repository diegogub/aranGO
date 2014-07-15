package aranGO

import(
  "reflect"
  "encoding/json"
)


func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}
	return val
}

func subParse(i map[string]interface{},doc interface{}) error{
  b,err := json.Marshal(i)
  if err != nil {
    return err
  }
  err = json.Unmarshal(b,doc)
  return err
}
