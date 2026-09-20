package model

import "encoding/json"

type SnowLumaSendGroupMessage struct {
	GroupID int64  `json:"group_id"`
	Message string `json:"message"`
}

type SnowLumaResponse struct {
	Status  string          `json:"status"`
	Retcode int64           `json:"retcode"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message,omitempty"`
	Wording string          `json:"wording,omitempty"`
}

type OB11Member struct {
	GroupID          int64  `json:"group_id"`
	UserID           int64  `json:"user_id"`
	Nickname         string `json:"nickname"`
	Card             string `json:"card"`
	IsRobot          bool   `json:"is_robot"`
	Sex              string `json:"sex"`
	Age              int    `json:"age"`
	JoinTime         int64  `json:"join_time"`
	LastSentTime     int64  `json:"last_sent_time"`
	ShutUpTimestamp  int64  `json:"shut_up_timestamp"`
	Level            string `json:"level"`
	Role             string `json:"role"`
	Title            string `json:"title"`
	Area             string `json:"area"`
	Unfriendly       bool   `json:"unfriendly"`
	TitleExpireTime  int64  `json:"title_expire_time"`
	CardChangeable   bool   `json:"card_changeable"`
	QidianMasterFlag int    `json:"qidian_master_flag"`
	QidianCrewFlag   int    `json:"qidian_crew_flag"`
	QidianCrewFlag2  int    `json:"qidian_crew_flag_2"`
}

type OB11Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type OB11TextData struct {
	Text string `json:"text"`
}

type OB11FaceData struct {
	ID string `json:"id"`
}

type OB11ImageData struct {
	File string `json:"file"`
}

type OB11RecordData struct {
	File string `json:"file"`
}

type OB11VideoData struct {
	File string `json:"file"`
}

type OB11AtData struct {
	QQ string `json:"qq"`
}

type OB11ReplyData struct {
	ID string `json:"id"`
}
