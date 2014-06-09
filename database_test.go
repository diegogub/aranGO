package aranGO

import (
	"fmt"
	"log"
	"testing"
	"time"
)

type TestDoc struct {
	Document
	Atr1 string `json:"at,omitempty" enum:"A,M,S"`
	Int  string `json:"int,omitempty" required:"-"`
	Sub  SubDoc `json:"doc" sub:"-"`
}

type SubDoc struct {
	Requi string `json:"q" required:"-"`
}

// implement modeler
func (t *TestDoc) GetKey() string {
	return t.Key
}

func (t *TestDoc) GetCollection() string {
	return "trans"
}

func (t *TestDoc) GetError() (string, bool) {
	return t.Message, t.Error
}

func (t *TestDoc) PreSave(err Error) {
	fmt.Println("presave hook")
	return
}

func (t *TestDoc) PostSave(err Error) {
	fmt.Println("postsave hook")
	return
}

func (t *TestDoc) PreUpdate(err Error) {
	fmt.Println("preupdate hook")
	return
}

func (t *TestDoc) PostUpdate(err Error) {
	fmt.Println("post update hook")
	return
}

func (t *TestDoc) PreDelete(err Error) {
	fmt.Println("pre delete hook")
	return
}

func (t *TestDoc) PostDelete(err Error) {
	fmt.Println("post delete hook")
	return
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

func Test_Database(t *testing.T) {
	/*
	  db ,err := NewDatabase("test","http://localhost:8529","diego","test")
	  log.Println(db)
	  db.Auth = true
	  db.Unsafe = true
	  db.Log = false

	  if db != nil {
	    err = db.Connect()
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
	Amount      int   `json:"c"`
	Commission  []int `json:"pcomm"`
	Clicks      []int `json:"clicks"`
	Impressions []int `json:"impre" `
}

func NewCounter() *Counter {
	var c Counter
	c.Commission = make([]int, 24)
	c.Clicks = make([]int, 24)
	c.Impressions = make([]int, 24)
	return &c
}

func Test_T(t *testing.T) {
	s, _ := Connect("http://localhost:8529", "diego", "test", false)
	db := s.DB("test2")
	var test TestDoc
	test.Atr1 = "A"
	test.Int = "test2"
	test.Sub.Requi = "miau"

	t0 := time.Now()
	log.Println(Save(db, &test))
	t1 := time.Now()
	log.Println(test)
	test.Atr1 = "M"

	log.Println(Save(db, &test))

	log.Println(t1.Sub(t0))
	/*
		log.Println(db)
		log.Println(db.Collections)
		//
		var tra Transaction
		co := NewCounter()
		co2 := NewCounter()
		co2.Key = "tet"
		co.Key = "counter"
		t1 := time.Now()
		db.Col("trans").Save(co2)
		t2 := time.Now()
		// Relate both counter
		//err := db.Col("rela").SaveEdge(map[string]interface{}{"pirchicho": "loco", "send": true}, "trans/tet", "trans/counter")
		//log.Print(err)
		//db.Col("trans").Delete("tet")

		var doc TestDoc
		doc.Int = 678

		tra.Collections = map[string][]string{}
		// Increase counter by 1
		// tra.Action = " function(params) { var db = require('internal').db ; var counter = db.trans.document( params.key) ; var change = {} ;var  m = counter[params.stat] ; m[params.hour] = m[params.hour] + params.inc; change[params.stat] = m ; db.trans.update(params.key, change); return m[params.hour]} "
		//tra.Params = map[string]interface{}{"key": "counter", "hour": 3, "stat": "pcomm", "inc": 2}
		// Sum whole
		tra.Action = " function(params) { var db = require('internal').db ; var counter = db.trans.document( params.key) ; var arr = counter[params.stat] ;var sum = 0 ; for (i=0;i< arr.length ; i++){ sum += arr[i]; } ; return sum; } "
		tra.Params = map[string]interface{}{"key": "counter", "hour": 3, "stat": "pcomm", "inc": 2}

		for i := 0; i < 1; i++ {
			db.ExecuteTran(&tra)
		}
		log.Println(tra.Result)
		log.Println(t2.Sub(t1))
	*/
}
