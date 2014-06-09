package aranGO

import (
	"errors"
	nap "github.com/jmcvetta/napping"
	"net/url"
)

type Session struct {
	host string
	nap  *nap.Session
	dbs  Databases
}

type Databases struct {
	List []string `json:"result" `
}

func Connect(host, user, password string, log bool) (*Session, error) {
	var sess Session
	var s nap.Session
	var dbs Databases
	var err error
	var request string
	s.Log = log
	// default unsafe
	s.UnsafeBasicAuth = true

	if user != "" {
		s.Userinfo = url.UserPassword(user, password)
	}

	request = host + "/_db/_system/_api/version"
	resp, err := s.Get(request, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	// check if 200
	switch resp.Status() {
	case 200:
		// load Databases
		request = host + "/_api/database/user"
		_, err = s.Get(request, nil, &dbs, nil)
		sess.dbs.List = dbs.List

		if err != nil {
			return nil, err
		}
		sess.nap = &s
		sess.host = host
		return &sess, nil
	default:
		return nil, errors.New("Invalid host or auth data to connect")
	}

}

func (s *Session) DB(name string) *Database {
	var db Database
	var found bool
	if name != "" {
		db.Name = name
	} else {
		return nil
	}

	for _, dbname := range s.dbs.List {
		if dbname == name {
			found = true
			break
		}
	}
	if found {
		db.baseURL = s.host + "/_db/" + db.Name + "/_api/"
		db.sess = s
		// load collections
		Collections(&db)
	} else {
		panic("Invalid DB")
	}

	return &db

}
