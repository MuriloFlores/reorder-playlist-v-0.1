package DTOs

type PlaylistActionDTO struct {
	ActionName string `json:"action_name"`
	PlaylistId string `json:"playlist_id"`
	Params     string `json:"params;omitempty"`
	Err        string `json:"err"`
	RetryAt    int64  `json:"retry_at"`
	UserId     string `json:"user_id"`
}
