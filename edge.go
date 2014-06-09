package aranGO

type Edge struct {
	Id   string `json:"_id"  `
	From string `json:"_from"`
	To   string `json:"_to"  `

	Error   bool   `json:"error,omitempty"`
	Message string `json:"errorMessage,omitempty"`
	Code    int    `json:"code,omitempty"`
	Num     int    `json:"errorNum,omitempty"`
}
