package aranGO

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

// Configure to start testing
var(
  collection = "apps"
  doc        DocTest
  dbname     = "goleat"
  username       = ""
  password   = ""
  verbose    = false
  server     = "http://localhost:8529"
  s *Session
)

func TestSimple(t *testing.T){
  // connect
  s ,err := Connect(server, username, password, verbose)
  assert.Nil(t,err)

  db := s.DB(dbname)
  assert.NotNil(t,db)

  c  := db.Col(collection)
  assert.NotNil(t,c)

  // Any
  err = c.Any(&doc)
  assert.Equal(t,doc.Error,false)
  assert.Nil(t,err)

  // Example
  cur, err := c.Example(map[string]interface{}{},0,10)
  assert.Equal(t,doc.Error,false)
  assert.Nil(t,err)
  assert.NotNil(t,cur)

  // need to add new functions!
}
