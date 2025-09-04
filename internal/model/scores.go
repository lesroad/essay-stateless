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
			All         int64 `json:"all"`
			Appearance  int64 `json:"appearance"`
			Content     int64 `json:"content"`
			Expression  int64 `json:"expression"`
			Structure   int64 `json:"structure"`
			Development int64 `json:"development"`
		} `json:"scores"`
	} `json:"result"`
}
