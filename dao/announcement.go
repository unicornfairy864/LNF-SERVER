package dao

import (
	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
)

type AnnouncementGroup struct{}

func (a *AnnouncementGroup) CreateAnnouncement(announcement *model.Announcement) error {
	err := global.LNF_DB.Create(announcement).Error
	return err
}

func (a *AnnouncementGroup) UpdateAnnouncement(announcement *model.Announcement) error {
	an := []string{"admin_id", "title", "content", "type", "status", "is_top", "view_count", "published_at", "created_at", "updated_at", "is_deleted"}
	err := global.LNF_DB.Model(&announcement).Select(an).Updates(&announcement).Error
	return err
}

// 取小于等于给定ID的公告
func (a *AnnouncementGroup) GetAnnouncementByID(id int64) *model.Announcement {
	var an *model.Announcement
	global.LNF_DB.Where("id <= ?", id).Last(&an)
	if an == nil || an.IsDeleted == 1 {
		return nil
	}
	return an
}
