package domain

type LogPassJson struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type AccessRefreshJson struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
