aranGO
======
~~~
go get github.com/diegogub/aranGO
~~~
Golang driver for ArangoDB.
It's under development, and I had not time to finish the documentation.

Here are the things you can do until now:

  * Databases : create
  * Collections : drop, create, list, truncate
  * Documents : save, replace,patch, query (simple query,AQL,Transactions)
  * Edges : Relate documents, save, patch, replace 
  * Execute transactions
  * Execute AQL

I'm planning to cover all functionalities after stable 2.2.0 realease, hopefully next week

Any ideas for the driver or bug fixes please feel free to create a issue or pull-request to dev :)

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
  var d1,d2 DocTest
  d1.Name = "Diego"
  d1.Age = 22
  d1.Likes = []string { "arangodb", "golang", "linux" }
  
  d2.Name = "Facundo"
  d2.Age = 25
  d2.Likes = []string { "php", "linux", "python" }


  err =s.DB("test").Col("docs1").Save(&d1)
  err =s.DB("test").Col("docs1").Save(&d2)
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
  d1.Age = 23
  err =s.DB("test").Col("docs1").Replace(d1.Key,d1)
  if err != nil {
    panic(err)
  }
  
  // Relate documents
  s.DB("test").Col("ed").Relate(d1.Key,d2.Key,map[string]interface{ "is" : "friend" })
~~~

AQL
===

~~~
// query query 
  q := ara.NewQuery("FOR i in docs1 RETURN i")
  c ,err:=s.DB("test").Execute(q)
  if err != nil {
    panic(err)
  }
  var doc DocTest

  for c.FetchOne(&doc){
    log.Println(doc)
  }

~~~

Transactions
===

~~~
// saving document with transaction
func TranSave(db *ara.Database,doc interface{},col string,counter string) (*ara.Transaction,error){
  if  col == "" || counter == ""{
    return nil,errors.New("Collection or counter must not be nil")
  }

  write := []string { col }
  q := `function(params){
                var db = require('internal').db;
                try {
                  var c = db.`+col+`.document('c');
                }catch(error){
                  var tmp = db.`+col+`.save( { '_key' : 'c' , '`+counter+`' : 0 });
                }
                var c = db.`+col+`.document('c');
                var co = c.`+counter+` || 0;
                co = co + 1 ;
                // update counter
                db.`+col+`.update(c, { '`+counter+`' : co }) ;
                params.doc.s = -1 * co ;
                params.doc.created = new Date().toISOString();
                var res = db.`+col+`.save(params.doc) ;
                return res._key
        }
  `
  t := ara.NewTransaction(q,write,nil)
  t.Params = map[string]interface{}{ "doc" : doc }

  err := t.Execute(db)

  return t,err
}
~~~
