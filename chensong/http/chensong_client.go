package http

type NapcatRequest struct {
	UserID    int64  `json:"user_id,omitempty"`
	GroupID   int64  `json:"group_id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
}

type ChenSongClient struct{}

func (c *ChenSongClient) Request(method string, napcatReq *NapcatRequest) ([]byte, error) {
}
