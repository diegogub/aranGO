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
  s,err := Connect("http://pampa.dev:8529","diego","test",true)
  if err != nil{
    panic(err)
  }
  db := s.DB("_system")

  if db == nil {
    panic("invalid db")
  }
  // test relations
  //ed := NewEdgeDefinition("edges",[]string{ "test" }, []string{ "test" })
  ed2 := NewEdgeDefinition("runs",[]string{ "test" }, []string{ "campaigns" })
  db.Graph("test1").AddEdge(ed2)

}
