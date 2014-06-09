package aranGO

import(
  "testing"
  "encoding/json"
  "log"
)


func Test_Cursor(t *testing.T){
  var c Cursor
  j := []byte( `{ "result" : [ 501, 509, 510 ], "hasMore" : true, "count" : 10, "extra" : { "fullCount" : 500 }, "error" : false,   "code" : 201 }`)

  json.Unmarshal(j,&c)

  log.Println("---",c)



}
