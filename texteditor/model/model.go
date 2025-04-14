package model

type Edit struct {
	User  string `json:"user"`
	Op    string `json:"op"`    // e.g., "insert" or "delete"
	Index int    `json:"index"` // position of the change
	Text  string `json:"text"`  // only for inserts
}
