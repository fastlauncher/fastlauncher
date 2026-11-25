package model

import (
	"encoding/json"
)

type App struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Keywords    string `json:"keywords"`
	Items       []App  `json:"items,omitempty"`  // inline submenu
	Config      string `json:"config,omitempty"` // reference to another json config
}

func NewAppListFromJSON(
	j []byte,
) (
	[]App,
	error,
) {
	response := []App{}

	if err := json.Unmarshal(j, &response); err != nil {
		return response, err
	}

	return response, nil
}
