package aranGO

type Transaction struct{
  Collections map[string][]string    `json:"collections"`
  Action      string                 `json:"action"`
  Result      interface{}            `json:"result,omitempty"`

  //Optional
  Sync        bool                    `json:"waitForSync,omitempty"`
  Lock        int                     `json:"lockTimeout,omitempty"`
  Replicate   bool                    `json:"replicate,omitempty"`
  Params      map[string]interface{}  `json:"params,omitempty"`

  //ErrorInfo
  Error       bool                    `json:"error,omitempty"`
  Code        int                     `json:"code,omitempty"`
  Num         int                     `json:"errorNum,omitempty"`
}
