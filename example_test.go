package aranGO

import "testing"
import "log"
import "time"

type Sub struct{
  SubAt  string `unique:"users"`
}

type DocTest struct {
  Document // arango Document to save id, key, rev
  Sub
  Name     string  `json:"" unique:"users"`
  Age      int
  Likes    []string
}

func (d *DocTest) GetCollection() string {
  return "test"
}

func (d *DocTest) GetKey() string {
  return d.Key
}

func (d *DocTest) GetError() (string,bool){
  return d.Message,d.Error
}

func TestExample(t *testing.T) {
  // connect
  log.Println("start test")
  s,err := Connect("http://localhost:8530","diego","test",true)
  if err != nil{
    panic(err)
  }
  db := s.DB("_system")

  d1 := DocTest{ Name : "Romina" }

  d2 := DocTest{ Name : "Diego" }
  d3 := DocTest { Name : "Emiliano" }

  c ,_:= NewContext(db)
  a ,e := c.NewRelation(&d1,Obj{ "in" : "love"},"relations","out",&d2,&d3)
  a.Update = false
  if len(e) > 0{
    panic(e)
  }
  t0:= time.Now()
  a.Commit()
  t1 := time.Now()
  log.Println(t1.Sub(t0))
  log.Println(d3)


  /*
  doc ,_:= NewDocument("persons/122636867")
  for i:=0 ; i<1000; i++ {
    m,err := doc.Map(db)
    log.Println(i,m)
    if err != nil {
        panic(err)
    }
  }
  c, err :=db.Col("persons").Indexes()
  log.Println(c)
  if err != nil {
    panic(err)
  }

  aq := NewAqlStruct()
  //log.Println(aq.For("u","users").For("adm","test").Filter(`{ "key" : "u" , "filters": [{ "name": "price", "op": "gt", "val": 12.12 },{ "name": "age", "op": "eq", "val": 21 }] , "any" : true}`).Return(Obj{ "u" : Atr("u","name") }).Generate())

  log.Println("----",Fil("name","eq",213).String(""))
  log.Println(aq.For("u","users").Let("test","hola").Filter(`{ "key" : "u" , "filters": [{ "name": "id", "op": "==", "field": "adm.id" },{ "name": "status", "op": "eq", "val":"A"}], "any" : true }`).Sort("u.name","u.test","ASC","u.age","DESC").Limit(2,10).Return(Obj{ "u" : Atr("u","name") }).Generate())
  */
}
