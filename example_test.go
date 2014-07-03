package aranGO

import "testing"
import "time"
import "log"

type Doc struct {
  Document
  Atr1 string  `json:"a1"`
  Num  int     `json:"num"`
  List int `json:"test"`
}

func TestExample(t *testing.T) {
  log.Println("Connection to database")

  s ,err:= Connect("http://pampa.dev:8529","diego","test",true)

  if err != nil {
    panic(err)
  }

  // Create Test Database
  user := []User{ User{ "diego","test",true,""}}
  err = s.CreateDB("pichicho",user)

  if err != nil {
    log.Println(err)
  }
  // create collection
  var colOpt CollectionOptions
  colOpt.Name = "test2"
  err = s.DB("pichicho").CreateCollection(&colOpt)
  var edges CollectionOptions
  edges.Name = "test2"
  edges.Type = 2
  if err != nil {
    log.Println(err)
  }
  col := s.DB("pichicho").Col("test2")
  if col == nil{
    panic("invalid colection")
  }
  // create document
  var doc Doc

  doc.Atr1 = "atributo 1"
  doc.Num  = 1
  col.Save(&doc)

  // update document
  doc.Atr1 = "e"
  doc.List = 28
  col.Replace(doc.Key,&doc)

  if doc.Error {
    log.Println(doc.Message)
  }

  err = col.Patch(doc.Key,map[string]int{ "num" : 99})
  if err != nil{
    panic(err)
  }

  db := s.DB("pichicho")

  //err = db.TruncateCollection("test2")

  q := NewQuery("FOR i in test2 FILTER i.test == 28 RETURN i")

  q.Batch = 5
  //q.Options["fullCount"] = true

  c, err := db.Execute(q)
  if err != nil{
    log.Println(err)
  }

  var i []Doc
  log.Println(c.FullCount())
  t1 := time.Now()
  c.FetchBatch(&i)
  log.Println(c.HasMore())
  log.Println(i)
  /*
  for more,err := c.FetchOne(&i); more ; more,err = c.FetchOne(&i){
    if err != nil {
      log.Println(err)
      break
    }
    log.Println(i)
  }
  */
  t2 := time.Now()
  log.Println("Time to fetch query:",t2.Sub(t1))
}
