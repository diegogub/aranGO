package aranGO

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

// Configure to start testing
var(
  TestCollection = "apps"
  TestDoc        DocTest
  TestDbName     = "goleat"
  TestUsername       = ""
  TestPassword   = ""
  verbose    = false
  TestServer     = "http://localhost:8529"
  s *Session
)

// document to test
type DocTest struct {
  Document // arango Document to save id, key, rev
}

func TestSimple(t *testing.T){
  // connect
  s ,err := Connect(TestServer, TestUsername, TestPassword, verbose)
  assert.Nil(t,err)

  db := s.DB(TestDbName)
  assert.NotNil(t,db)

  c  := db.Col(TestCollection)
  assert.NotNil(t,c)

  // Any
  err = c.Any(&TestDoc)
  assert.Equal(t,TestDoc.Error,false)
  assert.Nil(t,err)

  // Example
  cur, err := c.Example(map[string]interface{}{},0,10)
  assert.Equal(t,TestDoc.Error,false)
  assert.Nil(t,err)
  assert.NotNil(t,cur)

  // need to add new functions!
}
