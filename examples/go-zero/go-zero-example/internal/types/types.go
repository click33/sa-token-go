package types

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Device   string `json:"device,optional"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserInfoResponse struct {
	LoginID     string   `json:"loginId"`
	Permissions []string `json:"permissions"`
	Roles       []string `json:"roles"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
