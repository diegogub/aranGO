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
  s,err := Connect("http://localhost:8529","diego","test",false)
  if err != nil{
    panic(err)
  }
  db := s.DB("_system")

  // test relations
  var i interface{}
  i = db.Col("vertex").Count()
  log.Println(i)


}
