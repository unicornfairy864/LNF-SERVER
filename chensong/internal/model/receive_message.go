package model

type GroupMessageEvent struct {
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type"`
	SubType     string      `json:"sub_type"`
	MessageID   int64       `json:"message_id"`
	MessageSeq  int         `json:"message_seq"`
	GroupID     int64       `json:"group_id"`
	GroupName   string      `json:"group_name"`
	UserID      int64       `json:"user_id"`
	Message     string      `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Font        int         `json:"font"`
	Sender      Sender      `json:"sender"`
	Anonymous   interface{} `json:"anonymous"` // null 时为 nil，有值时为 map[string]interface{}
}

type Sender struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Card     string `json:"card"`
	Role     string `json:"role"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
}
