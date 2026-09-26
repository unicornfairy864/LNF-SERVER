package model

// UploadImageResponse 图片上传响应
type UploadImageResponse struct {
	// 图片相对 URL（如 /uploads/2026/09/26/{uuid}.jpg），前端展示时需拼接服务器源
	URL string `json:"url"`
}
