package entity

type UserSettings struct {
	Language      string `json:"language"`
	Timezone      string `json:"timezone"`
	Theme         string `json:"theme"`
	Notifications bool   `json:"notifications"`
}