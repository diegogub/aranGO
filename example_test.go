package aranGO

import "testing"
import "log"

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
  s,err := Connect("http://localhost:8529","diego","test",true)
  if err != nil{
    panic(err)
  }
  db := s.DB("_system")
  /*
  doc ,_:= NewDocument("persons/122636867")
  for i:=0 ; i<1000; i++ {
    m,err := doc.Map(db)
    log.Println(i,m)
    if err != nil {
        panic(err)
    }
  }
  */
  c, err :=db.Col("persons").Indexes()
  log.Println(c)
  if err != nil {
    panic(err)
  }
}
