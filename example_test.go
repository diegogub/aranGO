package aranGO

import "testing"
import "log"


type DocTest struct {
  Document // arango Document to save id, key, rev
  Name     string
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
  // create DB
  s.CreateDB("test",nil)

  // create Collections to test
  edges := NewCollectionOptions("edges",true)
  edges.IsEdge()
  log.Println(s.DB("test").ColExist("docs1"))
  if !s.DB("test").ColExist("docs1"){
    docs1 := NewCollectionOptions("docs1",true)
    s.DB("test").CreateCollection(docs1)
  }

  if !s.DB("test").ColExist("docs2"){
    docs2 := NewCollectionOptions("docs2",true)
    s.DB("test").CreateCollection(docs2)
  }

  if !s.DB("test").ColExist("ed"){
    edges := NewCollectionOptions("ed",true)
    edges.IsEdge()
    err = s.DB("test").CreateCollection(edges)
    if err != nil {
     panic(err)
    }
  }

  var d1 DocTest
  d1.Name = "Diego"
  d1.Age = 23
  d1.Likes = []string { "arangodb", "golang", "linux" }

  err =s.DB("test").Col("docs1").Save(&d1)
  if err != nil {
    panic(err)
  }

  // could also check error in document
  /*
  if d1.Error {
    panic(d1.Message)
  }
  */

  // update document
  d1.Age = 22
  err =s.DB("test").Col("docs1").Replace(d1.Key,d1)
  if err != nil {
    panic(err)
  }


  // fetch all documents with

  q := NewQuery("FOR i in docs1 RETURN i")
  c ,err:=s.DB("test").Execute(q)
  if err != nil {
    panic(err)
  }
  var doc DocTest

  for c.FetchOne(&doc){
    log.Println(doc)
  }


  // s.DB("test").TruncateCollection("docs1")

}
