package aranGO

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

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

	// Save
	var saveTestDoc DocTest
	saveTestDoc.Text = TestString
	err = c.Save(saveTestDoc)
	assert.Nil(t, err)

	// Any
	TestDoc = DocTest{} // Clean TestDoc variable
	err = c.Any(&TestDoc)
	assert.Equal(t, TestDoc.Error, false)
	assert.Equal(t, TestString, TestDoc.Text)

	// First
	TestDoc = DocTest{} // Clean TestDoc variable
	err = c.First(map[string]interface{}{"Text": TestString}, &TestDoc)
	assert.Equal(t, TestDoc.Error, false)
	assert.Equal(t, TestString, TestDoc.Text)

	// Example
	TestDoc = DocTest{} // Clean TestDoc variable
  cur, err := c.Example(map[string]interface{}{"Text" : TestString}, 0, 10)
	assert.Equal(t, TestDoc.Error, false)
	assert.Nil(t, err)
	assert.NotNil(t, cur)

	TestDoc = DocTest{} // Clean TestDoc variable
	moreFiles := cur.FetchOne(&TestDoc)
	assert.Equal(t, moreFiles, false)
	assert.Equal(t, TestString, TestDoc.Text)

	// need to add new functions!

}
