package client

import "encoding/json"

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
