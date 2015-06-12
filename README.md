aranGO
======
~~~
go get github.com/diegogub/aranGO
~~~
Golang driver for ArangoDB.

Here are the things you can do until now:

  * Databases : create
  * Collections : drop, create, list, truncate
  * Documents : save, replace,patch, query (simple query,AQL,Transactions)
  * Edges : Relate documents, save, patch, replace
  * Execute transactions
  * Execute AQL
  * Replication config

Additional Features
-------------------
  * Minimal Models with hooks
  * AqlBuilder ( https://gowalker.org/github.com/diegogub/aranGO#AqlStruct , check Filter, it has some nice JSON2AQL filter feature. If you have any suggestion about JSON format and new ideas to improve it feel free to write me or pull-request :P )

Any ideas for the driver or bug fixes please feel free to create a issue or pull-request to dev :)

Documentation
-------------

https://gowalker.org/github.com/diegogub/aranGO

Basic Usage
-----------
~~~~
import ara "github.com/diegogub/aranGO"

type DocTest struct {
  ara.Document // Must include arango Document in every struct you want to save id, key, rev after saving it
  Name     string
  Age      int
  Likes    []string
}
~~~~

Connect and create collections
-----------------------------------
~~~
    //change false to true if you want to see every http request
    //Connect(host, user, password string, log bool) (*Session, error) {
    s,err := ara.Connect("http://localhost:8529","diego","test",false) 
    if err != nil{
        panic(err)
    }

    // CreateDB(name string,users []User) error
    s.CreateDB("test",nil)

    // create Collections test if exist
    if !s.DB("test").ColExist("docs1"){
        // CollectionOptions has much more options, here we just define name , sync
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
        s.DB("test").CreateCollection(edges)
    }
~~~~

Create and Relate documents
---------------------------
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
  s.DB("test").Col("ed").Relate(d1.Id,d2.Id,map[string]interface{}{ "is" : "friend" })
~~~

AQL
---

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
------------

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

Models
------
To be a model, any struct must implement Modeler Interface.

~~~
type Modeler interface {
    // Returns current model key
    GetKey() string
    // Returns collection where I should save the model
    GetCollection() string
    // Error
    GetError() (string, bool)
}
~~~

Implement Modeler and add tags to struct..
~~~

type DocTest struct {
  ara.Document // Must include arango Document in every struct you want to save id, key, rev after saving it
// required tag for strings.
  Name     string `required:"-"`
// unique tag validate within collection users, if username is unique
  Username string `unique:"users"`
// enum tag checks string value
  Type     string `enum:"A,M,S"`
// Next release I will be implementing some other tags to validate int
and arrays
  Age      int
  Likes    []string
}

func (d *DocTest) GetKey() string{
  return d.Key
}

func (d *DocTest) GetCollection() string {
  return "testcollection"
}

func (d *DocTest) GetError()(string,error){
    // default error bool and messages. Could be any kind of error
    return d.Message,d.Error
}

// pass ArangoDB database as context
ctx, err :=  NewContext(db)

// save model, returns map of errors or empty map
e := ctx.Save(d1)

// check errors, also Error is saved in Context struct
if len(e) > 1 {
  panic(e)
}

// get other document
d2.Key = "d2key"
ctx.Get(d2)
log.Println(d2)

~~~


We can implement hooks to execute when saving,updating or deleting
model..
~~~
// execute before saving
func (d *DocTest) PreSave(c *ara.Context) {
   var e error
  // Any extra validation
  // ......
  // ......
  if e != nil {
    // errors should be set into context struct
    c.Err["presave"] = "failed to validate doctest"
  }

  return
}
~~~
