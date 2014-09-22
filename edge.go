package aranGO

type Edge struct {
	Id    string `json:"_id,omitempty"  `
	From  string `json:"_from"`
	To    string `json:"_to"  `
	Error bool   `json:"error,omitempty"`
	/*
	   Only vital info?
		Message string `json:"errorMessage,omitempty"`
		Code    int    `json:"code,omitempty"`
		Num     int    `json:"errorNum,omitempty"`
	*/
}
