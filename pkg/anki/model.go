package anki

import "encoding/json"

type ankiRequest struct {
	Action  string `json:"action"`
	Version int    `json:"version"`
	Params  any    `json:"params,omitempty"`
}

type ankiResponse struct {
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

type SyncResult struct {
	DeckName string
	Total    int
	Created  int
	Updated  int
}
