package models

type Country struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	DialCode    string `json:"dial_code"`
	FlagURL     string `json:"flag_url"`
}
