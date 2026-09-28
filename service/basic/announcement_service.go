package basic

import (
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/basic"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type AnnouncementServiceGroup struct{}

func (a *AnnouncementServiceGroup) CreateAnnouncement(anrq *model.AnnouncementUpdateRequest) response.Code {
	announcement := model.ToAnnouncement(anrq)
	announcement.ID = 0
	err := dao.AnnouncementDao.CreateAnnouncement(announcement).Error()
	if err != "" {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (a *AnnouncementServiceGroup) UpdatedAnnouncement(anrq *model.AnnouncementUpdateRequest) response.Code {
	announcement := model.ToAnnouncement(anrq)
	oldan := dao.AnnouncementDao.GetAnnouncementByID(announcement.ID)
	if oldan == nil {
		return response.CodeAnnouncementNotFound
	}
	if oldan.Status == 0 && announcement.Status == 1 {
		t := time.Now()
		announcement.PublishedAt = &t
	}
	err := dao.AnnouncementDao.UpdateAnnouncement(announcement).Error()
	if err != "" {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

func (a *AnnouncementServiceGroup) DeleteAnnouncement(id int64) response.Code {
	err := dao.AnnouncementDao.DeleteAnnouncement(id)
	if err != nil {
		return response.CodeDatabaseError
	}
	return response.CodeSuccess
}

// 从最新忽视ignore_pieces条后读取limit条，如无auth权限不返回草稿，已下架
func (a *AnnouncementServiceGroup) GetAnnouncements(anrq *model.AnnouncementGetRequest) *model.AnnouncementListResponse {
	announcements := []model.AnnouncementResponse{}
	ignore := anrq.IgnorePieces
	last := int64(999999)
	limit := anrq.Limit
	for limit > 0 && last > 0 {
		an := model.ToResponse(dao.AnnouncementDao.GetAnnouncementByID(last))
		if an == nil {
			return nil
		}
		status := an.Status
		if status == 0 && anrq.Auth && anrq.AdminID == an.AdminID || status == 2 && anrq.Auth || status == 1 {
			if ignore == 0 {
				announcements = append(announcements, *an)
				limit--
			} else {
				ignore--
			}
		}
		last = an.ID - 1
	}
	total, err := dao.AnnouncementDao.GetAmmount()
	if err != nil {
		return model.ToListResponse(0, announcements)
	}
	return model.ToListResponse(total, announcements)
}
