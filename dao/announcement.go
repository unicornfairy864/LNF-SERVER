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

func (a *AnnouncementGroup) DeleteAnnouncement(id int64) error {
	err := global.LNF_DB.Delete(id).Error
	return err
}

func (a *AnnouncementGroup) GetAmmount() (int64, error) {
	var count int64
	err := global.LNF_DB.Count(&count).Error
	return count, err
}

// 取小于等于给定ID的公告
func (a *AnnouncementGroup) GetAnnouncementByID(id int64) *model.Announcement {
	var an model.Announcement
	err := global.LNF_DB.Where("id <= ?", id).Last(&an)
	if err != nil || an.IsDeleted == 1 {
		return nil
	}
	return &an
}
