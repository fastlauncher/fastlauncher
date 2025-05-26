package model

import (
	"encoding/json"
)

type App struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Keywords    string `json:"keywords"`
}

func NewAppListFromJson(
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

func (a *App) ToJson() (
	[]byte,
	error,
) {
	return json.Marshal(a)
}
