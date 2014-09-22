package aranGO

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Configure to start testing
var (
	TestCollection = "apps"
	TestDoc        DocTest
	TestDbName     = "goleat"
	TestUsername   = ""
	TestPassword   = ""
	verbose        = false
	TestServer     = "http://localhost:8529"
	s              *Session
)

// document to test
type DocTest struct {
	Document // arango Document to save id, key, rev
	Text     string
}

func TestSimple(t *testing.T) {
	// connect
	s, err := Connect(TestServer, TestUsername, TestPassword, verbose)
	assert.Nil(t, err)

	// Create the db
	s.CreateDB(TestDbName, nil)
	defer s.DropDB(TestDbName)

	db := s.DB(TestDbName)
	assert.NotNil(t, db)

	c := db.Col(TestCollection)
	assert.NotNil(t, c)

	// Any
	err = c.Any(&TestDoc)
	assert.Equal(t, TestDoc.Error, false)
	assert.Nil(t, err)

	// Save
	var saveTestDoc DocTest
	saveTestDoc.Text = "Stringy string"
	err = c.Save(saveTestDoc)
	assert.Nil(t, err)

	// Example
	cur, err := c.Example(map[string]interface{}{}, 0, 10)
	assert.Equal(t, TestDoc.Error, false)
	assert.Nil(t, err)
	assert.NotNil(t, cur)

	var newTestDoc DocTest
	moreFiles := cur.FetchOne(&newTestDoc)
	assert.Equal(t, moreFiles, false)
	assert.Equal(t, saveTestDoc.Text, newTestDoc.Text)

	// need to add new functions!

}
