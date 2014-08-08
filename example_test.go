package aranGO

import "testing"
import "log"
import "encoding/json"

type Sub struct{
  SubAt  string `unique:"users"`
}

type DocTest struct {
  Document // arango Document to save id, key, rev
  Sub
  Name     string  `json:"-,omitempty" unique:"users"`
  Age      int
  Likes    []string
}

func TestExample(t *testing.T) {
  // connect
  log.Println("start test")
  s,err := Connect("http://pampa.dev:8529","diego","test",false)
  if err != nil{
    panic(err)
  }
  db := s.DB("pampa")

  col := db.Col("users")  
  uniq ,_:= col.Unique("Username","probando",true,"")

  log.Println(uniq)

  var test DocTest
  var e Error
  e = make(map[string]string)

  Unique(test,db,true,e)

  if db == nil {
    panic("invalid db")
  }

  var q AqlStruct

  type F struct {
    Filters []Filters `json:"filters"`
  }
  var f F

  post := `{ "filters" : [  {"conditions" : [ { "obj": "ops" , "op":"==" , "val":"1" },{ "obj": "type" , "op":"==" , "val":"CPM" }]}]    }`

  err = json.Unmarshal([]byte(post),&f)
  if err != nil {
    panic(err)
  }
  log.Println(f)
  q.For("acc").In("accounts")
  q.Filter("acc",f.Filters[0].Filter)
//q.Return(map[string]interface{}{ "acc" : "acc"})
  q.Custom("RETURN acc")


  w :=  []string{ "Pot","Jon","snow"}
  in := []string{ "name", "surname" }
  c := "test1"

  text := FullText(w,in,c)
  log.Println(text)

}
