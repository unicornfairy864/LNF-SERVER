package dao

import (
	"context"
	"errors"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/global"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"gorm.io/gorm"
)

type AdminStatsGroup struct{ Context context.Context }

func (g *AdminStatsGroup) db() *gorm.DB {
	if g.Context != nil {
		return global.LNF_DB.WithContext(g.Context)
	}
	return global.LNF_DB
}

// GetOverviewByCreatedPeriod 统计 created_at 落在同一周期内的物品当前快照。
func (g *AdminStatsGroup) GetOverviewByCreatedPeriod(start, end, pendingDeadline time.Time) (model.StatsOverviewRow, int64, error) {
	var row model.StatsOverviewRow
	var pendingOver24h int64
	err := g.db().Table("items").
		Select(`
			COUNT(*) AS published,
			COALESCE(SUM(CASE WHEN status = 2 AND claim_user_id IS NOT NULL AND claim_time IS NOT NULL THEN 1 ELSE 0 END), 0) AS returned,
			COALESCE(SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END), 0) AS pending,
			COALESCE(SUM(CASE WHEN status = 1 AND claim_user_id IS NOT NULL THEN 1 ELSE 0 END), 0) AS claimed,
			COALESCE(SUM(CASE WHEN status = 2 AND claim_user_id IS NULL THEN 1 ELSE 0 END), 0) AS closed`).
		Where("is_deleted = 0 AND created_at >= ? AND created_at < ?", start, end).
		Scan(&row).Error
	if err != nil {
		return row, 0, err
	}
	err = g.db().Table("items").
		Where("is_deleted = 0 AND status = 0 AND created_at >= ? AND created_at < ? AND created_at < ?",
			start, end, pendingDeadline,
		).Count(&pendingOver24h).Error
	return row, pendingOver24h, err
}

func (g *AdminStatsGroup) GetTrendByCreatedDate(start, end time.Time) ([]model.StatsTrendRow, error) {
	var rows []model.StatsTrendRow
	err := g.db().Table("items").
		Select(`DATE(created_at) AS date,
			COUNT(*) AS published,
			COALESCE(SUM(CASE WHEN status = 2 AND claim_user_id IS NOT NULL AND claim_time IS NOT NULL THEN 1 ELSE 0 END), 0) AS returned`).
		Where("is_deleted = 0 AND created_at >= ? AND created_at < ?", start, end).
		Group("DATE(created_at)").Order("date ASC").Scan(&rows).Error
	return rows, err
}

func (g *AdminStatsGroup) GetLocationsByCreatedPeriod(start, end time.Time, limit int) ([]model.StatsLocationRow, int64, error) {
	var rows []model.StatsLocationRow
	query := g.db().Table("items AS i").
		Select(`i.location_id, COALESCE(l.name, '') AS name, COUNT(*) AS count`).
		Joins("LEFT JOIN locations AS l ON l.id = i.location_id").
		Where("i.is_deleted = 0 AND i.created_at >= ? AND i.created_at < ?", start, end).
		Group("i.location_id, l.name").Order("count DESC, i.location_id ASC")
	if limit == 0 {
		query = query.Limit(1002)
	} else {
		query = query.Where("i.location_id IS NOT NULL").Limit(limit)
	}
	err := query.Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	if statsLocationLimitExceeded(rows, limit) {
		return nil, 0, ErrStatsLocationLimit
	}
	if limit > 0 {
		var unknown model.StatsLocationRow
		err = g.db().Table("items").Select("COUNT(*) AS count").Where("is_deleted = 0 AND location_id IS NULL AND created_at >= ? AND created_at < ?", start, end).Scan(&unknown).Error
		if err != nil {
			return nil, 0, err
		}
		unknown.Name = "unknown"
		rows = append(rows, unknown)
	}
	var total int64
	err = g.db().Table("items").Where("is_deleted = 0 AND created_at >= ? AND created_at < ?", start, end).Count(&total).Error
	return rows, total, err
}

var ErrStatsLocationLimit = errors.New("statistics location bucket limit exceeded")
var ErrStatsTagLimit = errors.New("statistics tag bucket limit exceeded")

func statsLocationLimitExceeded(rows []model.StatsLocationRow, limit int) bool {
	if limit != 0 {
		return false
	}
	known := 0
	for _, row := range rows {
		if row.LocationID != nil {
			known++
		}
	}
	return known > 1000
}

func statsTagLimitExceeded(rows []model.StatsDistributionRow) bool {
	known := 0
	for _, row := range rows {
		if row.BucketID != nil {
			known++
		}
	}
	return known > 1000
}

func (g *AdminStatsGroup) GetStatsItems(q model.StatsItemsQuery, now, deadline time.Time, highView bool) ([]model.StatsItemRow, int64, error) {
	query := g.db().Table("items AS i").Where("i.is_deleted = 0 AND i.status = 0 AND i.created_at <= ?", deadline)
	if highView {
		query = query.Where("i.view_count >= ?", q.MinViews)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "i.created_at ASC, i.id ASC"
	if highView {
		order = "i.view_count DESC, i.created_at ASC, i.id ASC"
	}
	rows := make([]model.StatsItemRow, 0)
	err := query.Select("i.id, i.title, i.type, i.status, i.location_id, COALESCE(l.name, '') AS location_name, i.view_count, i.created_at, TIMESTAMPDIFF(DAY, i.created_at, ?) AS stagnant_days", now).
		Joins("LEFT JOIN locations AS l ON l.id = i.location_id").Order(order).Offset((q.Page - 1) * q.PageSize).Limit(q.PageSize).Scan(&rows).Error
	return rows, total, err
}

func statsDurationSQL(groupBy string) (string, bool) {
	column := "NULL"
	switch groupBy {
	case "none":
	case "location":
		column = "location_id"
	case "type":
		column = "type"
	default:
		return "", false
	}
	// Only fixed whitelist columns may enter this SQL expression.
	return `WITH samples AS (
 SELECT ` + column + ` AS group_id, TIMESTAMPDIFF(SECOND, created_at, updated_at) AS duration
 FROM items WHERE is_deleted=0 AND status=2 AND claim_user_id IS NOT NULL AND claim_time IS NOT NULL
 AND updated_at >= ? AND updated_at < ? AND updated_at >= created_at
 ), ranked AS (
 SELECT group_id, duration, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY duration) AS rn,
 COUNT(*) OVER (PARTITION BY group_id) AS n FROM samples
 )
 SELECT group_id, COUNT(*) AS count, ROUND(AVG(duration),2) AS average_seconds,
 ROUND(AVG(CASE WHEN rn IN (FLOOR((n+1)/2), FLOOR((n+2)/2)) THEN duration END),2) AS median_seconds
 FROM ranked GROUP BY group_id ORDER BY group_id ASC`, true
}

func scanStatsRows[T any](db *gorm.DB, target *[]T) error {
	rows, err := db.Rows()
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var row T
		if err := db.ScanRows(rows, &row); err != nil {
			return err
		}
		*target = append(*target, row)
	}
	return rows.Err()
}

func (g *AdminStatsGroup) GetReturnDuration(start, end time.Time, groupBy string) (model.StatsDurationRow, []model.StatsDurationRow, error) {
	var overall model.StatsDurationRow
	groups := make([]model.StatsDurationRow, 0)
	query, ok := statsDurationSQL(groupBy)
	if !ok {
		return overall, groups, errors.New("invalid statistics duration dimension")
	}
	allSQL, _ := statsDurationSQL("none")
	var all []model.StatsDurationRow
	if err := g.db().Raw(allSQL, start, end).Scan(&all).Error; err != nil {
		return overall, groups, err
	}
	if len(all) > 0 {
		overall = all[0]
	}
	if groupBy != "none" {
		var err error
		if g.Context != nil {
			query += " LIMIT 10000"
			err = scanStatsRows(g.db().Raw(query, start, end), &groups)
		} else {
			err = g.db().Raw(query, start, end).Scan(&groups).Error
		}
		if err != nil {
			return overall, groups, err
		}
	}
	return overall, groups, nil
}

func (g *AdminStatsGroup) GetTimeHeatmap(start, end time.Time) ([]model.StatsHeatmapRow, error) {
	var rows []model.StatsHeatmapRow
	err := g.db().Table("items").Select("WEEKDAY(created_at) AS weekday, HOUR(created_at) AS hour, COUNT(*) AS count").Where("is_deleted = 0 AND created_at >= ? AND created_at < ?", start, end).Group("WEEKDAY(created_at), HOUR(created_at)").Scan(&rows).Error
	return rows, err
}

func statsDistributionSQL(dimension string) (string, bool) {
	switch dimension {
	case "type":
		return `SELECT i.type AS bucket_id, CASE i.type WHEN 0 THEN 'lost' WHEN 1 THEN 'found' ELSE '' END AS bucket_name,
 COUNT(*) AS published, SUM(CASE WHEN i.status=2 AND i.claim_user_id IS NOT NULL AND i.claim_time IS NOT NULL THEN 1 ELSE 0 END) AS returned
 FROM items i WHERE i.is_deleted=0 AND i.created_at >= ? AND i.created_at < ?
 GROUP BY i.type ORDER BY published DESC, bucket_id ASC`, true
	case "tag":
		return `SELECT it.tag_id AS bucket_id, CASE WHEN it.tag_id IS NULL THEN 'untagged' ELSE COALESCE(t.name,'') END AS bucket_name,
 COUNT(DISTINCT i.id) AS published, COUNT(DISTINCT CASE WHEN i.status=2 AND i.claim_user_id IS NOT NULL AND i.claim_time IS NOT NULL THEN i.id END) AS returned
 FROM items i LEFT JOIN item_tags it ON it.item_id=i.id LEFT JOIN tags t ON t.id=it.tag_id
 WHERE i.is_deleted=0 AND i.created_at >= ? AND i.created_at < ?
 GROUP BY it.tag_id,t.name ORDER BY published DESC,bucket_id ASC LIMIT 1002`, true
	default:
		return "", false
	}
}

func (g *AdminStatsGroup) GetDistribution(start, end time.Time, dimension string) ([]model.StatsDistributionRow, int64, error) {
	query, ok := statsDistributionSQL(dimension)
	if !ok {
		return nil, 0, errors.New("invalid statistics dimension")
	}
	rows := make([]model.StatsDistributionRow, 0)
	var err error
	if g.Context != nil {
		err = scanStatsRows(g.db().Raw(query, start, end), &rows)
	} else {
		err = g.db().Raw(query, start, end).Scan(&rows).Error
	}
	if err != nil {
		return nil, 0, err
	}
	if dimension == "tag" && statsTagLimitExceeded(rows) {
		return nil, 0, ErrStatsTagLimit
	}
	var total int64
	err = g.db().Table("items").Where("is_deleted=0 AND created_at >= ? AND created_at < ?", start, end).Count(&total).Error
	return rows, total, err
}

const statsFunnelSQL = `SELECT
 (SELECT COUNT(*) FROM users WHERE is_deleted=0 AND created_at >= ? AND created_at < ?) AS registered,
 COUNT(DISTINCT CASE WHEN created_at >= ? AND created_at < ? THEN user_id END) AS published,
 COUNT(DISTINCT CASE WHEN claim_time >= ? AND claim_time < ? THEN claim_user_id END) AS claimed,
 COUNT(DISTINCT CASE WHEN status=2 AND claim_time IS NOT NULL AND updated_at >= ? AND updated_at < ? THEN claim_user_id END) AS returned
 FROM items WHERE is_deleted=0 AND ((created_at >= ? AND created_at < ?) OR (claim_time >= ? AND claim_time < ?) OR (updated_at >= ? AND updated_at < ?))`

func (g *AdminStatsGroup) GetFunnel(start, end time.Time) ([]model.StatsFunnelRow, error) {
	var row struct{ Registered, Published, Claimed, Returned int64 }
	err := g.db().Raw(statsFunnelSQL, start, end, start, end, start, end, start, end, start, end, start, end, start, end).Scan(&row).Error
	return []model.StatsFunnelRow{{Stage: "registered", Users: row.Registered}, {Stage: "published", Users: row.Published}, {Stage: "claimed", Users: row.Claimed}, {Stage: "returned", Users: row.Returned}}, err
}
