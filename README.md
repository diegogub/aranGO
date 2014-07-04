aranGO
======

Golang driver for ArangoDB


Basic Usage
===========
~~~~
import ara "github.com/diegogub/aranGO"

type DocTest struct {
  ara.Document // Must include arango Document in every struct you want to save id, key, rev after saving it
  Name     string
  Age      int
  Likes    []string
}
~~~~
~~~
 // Connecting to arangoDB
 // change false to true if you want to see every http request
 s,err := ara.Connect("http://localhost:8529","diego","test",false) 

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
    edges.IsEdge() // set to Edge
    err = s.DB("test").CreateCollection(edges)
    if err != nil {
     panic(err)
    }
  }
~~~~

~~~
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
~~~
