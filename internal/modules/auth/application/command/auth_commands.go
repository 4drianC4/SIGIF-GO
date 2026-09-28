package command

type LoginCommand struct {
	Email      string
	Password   string
	UserAgent  string
	IPAddress  string
}

type RefreshCommand struct {
	RefreshToken string
}

type LogoutCommand struct {
	RefreshToken string
}

type LogoutAllCommand struct {
	UserID string
}