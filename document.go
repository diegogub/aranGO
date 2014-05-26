package aranGO

import(
)

type Document struct {
  Id    string         `json:"_id,omitempty"  `
  Rev   string         `json:"_rev,omitempty" `
  Key   string         `json:"_key,omitempty" `

  Message    string     `json:"errorMessage,omitempty"`
  Code       int        `json:"code,omitempty"`
  Num        int        `json:"errorNum,omitempty"`
}

func (d *Document) SetKey(key string) error{
  //valitated key
  d.Key = key
  return nil
}

func (d *Document) SetRev(rev string) error{
  d.Rev = rev
  return nil
}
