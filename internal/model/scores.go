package model

type APIScore struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Result  struct {
		Comment  string `json:"comment"`
		Comments struct {
			Appearance  string `json:"appearance"`
			Content     string `json:"content"`
			Expression  string `json:"expression"`
			Structure   string `json:"structure"`
			Development string `json:"development"`
		} `json:"comments"`
		Scores struct {
			All         int `json:"all"`
			Appearance  int `json:"appearance"`
			Content     int `json:"content"`
			Expression  int `json:"expression"`
			Structure   int `json:"structure"`
			Development int `json:"development"`
		} `json:"scores"`
	} `json:"result"`
}
