package domain

type LogPassJSON struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AccessRefreshJSON struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
