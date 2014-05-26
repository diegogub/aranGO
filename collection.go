package aranGO


// TODO Must Implement revision control

import(
  //nap "github.com/jmcvetta/napping"
  "log"
  "reflect"
  "errors"
)


type Collection struct {
  db   *Database `json:"db"`
  name string    `json:"name"`

  // last ,
  policy     string `json:"-"`
  revision   bool   `json:"-"`
}

// Save saves doc into collection, doc should have Document Embedded to retrieve error and Key later.
func (col *Collection) Save(doc interface{}) error{
  // Get URL to post
    res ,err :=col.db.send("document?collection=" + col.name ,"","POST",doc,&doc,&doc)

    if err != nil{
      return err
    }

    if res.Status() != 200{
      return errors.New("Unable to save document error")
    }

    return nil
}

//Get Document
func (col *Collection) Get(key string,doc interface{}) error{
  if key == ""{
    return errors.New("Key must not be empty")
  }

  //var ErrorDoc ErrorDocument
  _, err := col.db.get("document",col.name + "/" + key,"GET",nil,&doc,&doc)

  if err != nil {
    return err
  }

  return nil
}

// Replace document
func (col *Collection) Replace(key string, doc interface{}) error{
  if key == ""{
    return errors.New("Key must not be empty")
  }

  resp,err :=col.db.send("document",col.name + "/" + key,"PUT",doc,&doc,&doc)

  if err != nil{
    return err
  }

  if resp.Status() != 201 {
    return errors.New("Unable to replace document")
  }

  return nil
}

func (col *Collection) Patch(key string,doc interface{}) error{
  if key == ""{
    return errors.New("Key must not be empty")
  }

  resp,err :=col.db.send("document",col.name + "/" + key,"PATCH",doc,&doc,&doc)

  if err != nil{
    return err
  }

  if resp.Status() != 201 {
    return errors.New("Unable to replace document")
  }

  return nil
}

func (col *Collection) Delete(key string) error{
  if key == ""{
    return errors.New("Key must not be empty")
  }

  resp,err :=col.db.get("document",col.name + "/" + key,"DELETE",nil,nil,nil)
  if err != nil{
    return err
  }
  log.Println(resp)
  return nil
  switch resp.Status(){
    case 202, 200:
      return nil
    default:
      return errors.New("Document don't exist or revision error")

  }
}


func (col *Collection) Exist() bool{
  return true
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
