package aranGO

import(
  "testing"
  "log"
)


type TestDoc struct{
  Document
  Atr1       string     `json:"at"`
  Int        int        `json:"int"`
  Array      []string   `json:"ar"`
}

/*
func (t *TestDoc) SetDoc(d Document) {
  t.Id = d.Id
  t.Rev = d.Rev
  t.Key = d.Key
  return
}

func (t *TestDoc) SetError(d ErrorDocument){
  var er ErrorDocument
  er = d
  t.Err = &er
  return
}
*/

func Test_Database(t *testing.T){
  /*
  db ,err := NewDatabase("test","http://localhost:8529","diego","test")
  log.Println(db)
  db.Auth = true
  db.Unsafe = true
  db.Log = false

  if db != nil {
    err = db.Ping()
    log.Println(err)
  }

  doc := new(TestDoc)

  err = db.Col("test").Get("2170700199",doc)

  log.Println(doc)
  p :=new(TestDoc)
  p.Int = 14

  err = db.Col("test").Patch(doc.Key,p)
  log.Println("document after replace:",doc)
  log.Println("error:",err)
  return 

  /*
  for i:=0 ; i <= 20 ; i++{

    doc := new(TestDoc)
    //doc.Key = "peterete"

    doc.Int = 25
    doc.Atr1 = "diego"
    doc.Array = make([]string,10)
    doc.Array[0] = "test"
    doc.Array[3] = "miau"

    err = db.Col("test").Save(doc)
  }
  q := NewQuery("FOR i in test FILTER i.int == 14 RETURN i")
  q.Batch = 2
  log.Println(q)
  q.Count = true
  q.Validate = true

  c , err:= db.Execute(q)
  log.Println("query executed")

  if err != nil {
    log.Println("cursor error",err)
  }else{
  te := new(TestDoc)

  ok, err := c.Next(te)
  log.Println(te,ok,c.Index)
  for ok {
    ok, err = c.Next(te)
    log.Println(te,ok,c.Index)
    if err != nil{
      log.Println(err)
    }
  }
  }
  err = db.Col("test").Delete(doc.Key)
  log.Println(err)
  */
}

type Counter struct {
  Document
  Amount int   `json:"c"`
}

func Test_T(t *testing.T){
  db ,_:= NewDatabase("test","http://localhost:8529","diego","test")
  log.Println(db)
  db.Auth = true
  db.Unsafe = true
  db.Log = true
  db.Ping()
//
  var tra Transaction
  var co  Counter
  co.Key = "counter"
  var s   bool

  db.Col("trans").Save(co)

  var doc TestDoc
  doc.Int = 678

  tra.Collections = map[string][]string { "write" : []string { "trans" } }
  tra.Action = " function(params) { var db = require('internal').db ; var counter = db.trans.document( params.key) ; var v = counter[params.skey] ; v = v + params.inc ; db.trans.update(params; return v} "


  tra.Params = map[string]interface{} { "key" : "counter" , "skey" : "c", "inc" : 1 }

  db.ExecuteTran(&tra,&s)
  log.Println(tra.Result)
}
