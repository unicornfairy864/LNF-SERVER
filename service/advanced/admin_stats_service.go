package advanced

import (
	"context"
	"errors"
	"math"
	"math/big"
	"time"

	"github.com/unicornfairy864/LNF-SERVER/dao"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

type AdminStatsServiceGroup struct{ Context context.Context }

func (s *AdminStatsServiceGroup) statsDAO() *dao.AdminStatsGroup {
	return &dao.AdminStatsGroup{Context: s.Context}
}

var AdminStatsService AdminStatsServiceGroup

var statsLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("CST", 8*60*60)
	}
	return loc
}()

func parseStatsPeriod(q *model.StatsDateQuery, now time.Time, defaultDays int) (model.StatsPeriod, response.Code) {
	loc := statsLocation
	now = now.In(loc)
	hasDates := q.StartDate != "" || q.EndDate != ""
	if hasDates && (q.StartDate == "" || q.EndDate == "" || q.Days != 0) {
		return model.StatsPeriod{}, response.CodeParamError
	}
	if !hasDates {
		days := q.Days
		if days == 0 {
			days = defaultDays
		}
		if days < 1 || days > 366 {
			return model.StatsPeriod{}, response.CodeParamError
		}
		end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
		start := end.AddDate(0, 0, -days)
		return model.StatsPeriod{StartDate: start.Format("2006-01-02"), EndDate: end.AddDate(0, 0, -1).Format("2006-01-02"), Start: start, End: end}, response.CodeSuccess
	}
	start, err1 := time.ParseInLocation("2006-01-02", q.StartDate, loc)
	endDate, err2 := time.ParseInLocation("2006-01-02", q.EndDate, loc)
	if err1 != nil || err2 != nil || start.After(endDate) {
		return model.StatsPeriod{}, response.CodeParamError
	}
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc)
	end := endDate.AddDate(0, 0, 1)
	if end.After(tomorrow) || end.Sub(start) > 366*24*time.Hour {
		return model.StatsPeriod{}, response.CodeParamError
	}
	return model.StatsPeriod{StartDate: q.StartDate, EndDate: q.EndDate, Start: start, End: end}, response.CodeSuccess
}

func metric(value, previous int64) model.StatsMetric {
	result := model.StatsMetric{Value: value, Previous: previous, Comparable: true}
	if previous == 0 {
		if value == 0 {
			result.ChangePercent = floatPtr(0)
		} else {
			result.Comparable = false
		}
		return result
	}
	result.ChangePercent = floatPtr(math.Round(float64(value-previous)*10000/float64(previous)) / 100)
	return result
}

func rateMetric(value, previous float64) model.StatsRateMetric {
	result := model.StatsRateMetric{Value: value, Previous: previous, Comparable: true}
	if previous == 0 {
		if value == 0 {
			result.ChangePercent = floatPtr(0)
		} else {
			result.Comparable = false
		}
		return result
	}
	result.ChangePercent = floatPtr(math.Round((value-previous)*10000/previous) / 100)
	return result
}

func floatPtr(v float64) *float64 { return &v }

func periodBefore(p model.StatsPeriod) model.StatsPeriod {
	days := int(p.End.Sub(p.Start).Hours() / 24)
	return model.StatsPeriod{Start: p.Start.AddDate(0, 0, -days), End: p.Start, StartDate: p.Start.AddDate(0, 0, -days).Format("2006-01-02"), EndDate: p.Start.AddDate(0, 0, -1).Format("2006-01-02")}
}

func overviewPeriods(q *model.StatsDateQuery, now time.Time) (model.StatsPeriod, model.StatsPeriod, response.Code) {
	if q.StartDate != "" || q.EndDate != "" || q.Days != 0 {
		current, code := parseStatsPeriod(q, now, 30)
		if code != response.CodeSuccess {
			return model.StatsPeriod{}, model.StatsPeriod{}, code
		}
		return current, periodBefore(current), response.CodeSuccess
	}

	now = now.In(statsLocation)
	current := model.StatsPeriod{
		Start: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, statsLocation),
		End:   time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, statsLocation),
	}
	current.StartDate = current.Start.Format("2006-01-02")
	current.EndDate = current.End.AddDate(0, 0, -1).Format("2006-01-02")
	previous := model.StatsPeriod{Start: current.Start.AddDate(0, -1, 0), End: current.Start}
	previous.StartDate = previous.Start.Format("2006-01-02")
	previous.EndDate = previous.End.AddDate(0, 0, -1).Format("2006-01-02")
	return current, previous, response.CodeSuccess
}

func buildTrendResponse(p model.StatsPeriod, rows []model.StatsTrendRow) *model.TrendResponse {
	byDate := make(map[string]model.StatsTrendRow, len(rows))
	for _, row := range rows {
		byDate[row.Date.In(statsLocation).Format("2006-01-02")] = row
	}
	points := make([]model.TrendPoint, 0, int(p.End.Sub(p.Start).Hours()/24))
	for day := p.Start; day.Before(p.End); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		row := byDate[key]
		points = append(points, model.TrendPoint{Date: key, Published: row.Published, Returned: row.Returned})
	}
	return &model.TrendResponse{StartDate: p.StartDate, EndDate: p.EndDate, Points: points}
}

func buildLocationsResponse(rows []model.StatsLocationRow, total int64, limit int) *model.LocationsResponse {
	unknown := model.LocationStat{Name: "unknown"}
	locations := make([]model.LocationStat, 0, len(rows))
	for _, row := range rows {
		percent := float64(0)
		if total > 0 {
			percent = math.Round(float64(row.Count)*10000/float64(total)) / 100
		}
		stat := model.LocationStat{LocationID: row.LocationID, Name: row.Name, Count: row.Count, Percent: percent}
		if row.LocationID == nil {
			unknown = stat
			continue
		}
		locations = append(locations, stat)
	}
	if limit > 0 && len(locations) > limit {
		locations = locations[:limit]
	}
	return &model.LocationsResponse{Total: total, Locations: locations, Unknown: unknown}
}

func (s *AdminStatsServiceGroup) Overview(q *model.StatsDateQuery, now time.Time) (*model.OverviewResponse, response.Code) {
	p, prev, code := overviewPeriods(q, now)
	if code != response.CodeSuccess {
		return nil, code
	}
	current, pending24, err := s.statsDAO().GetOverviewByCreatedPeriod(p.Start, p.End, now.In(statsLocation).Add(-24*time.Hour))
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	old, _, err := s.statsDAO().GetOverviewByCreatedPeriod(prev.Start, prev.End, prev.End.Add(-24*time.Hour))
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	currentTotal := current.Returned + current.Pending + current.Claimed + current.Closed
	oldTotal := old.Returned + old.Pending + old.Claimed + old.Closed
	currentRate, oldRate := float64(0), float64(0)
	if currentTotal > 0 {
		currentRate = math.Round(float64(current.Returned)*10000/float64(currentTotal)) / 100
	}
	if oldTotal > 0 {
		oldRate = math.Round(float64(old.Returned)*10000/float64(oldTotal)) / 100
	}
	return &model.OverviewResponse{Period: p, Published: metric(current.Published, old.Published), Returned: metric(current.Returned, old.Returned), Pending: metric(current.Pending, old.Pending), ReturnRate: rateMetric(currentRate, oldRate), PendingOver24h: pending24}, response.CodeSuccess
}

func (s *AdminStatsServiceGroup) Trend(q *model.StatsDateQuery, now time.Time) (*model.TrendResponse, response.Code) {
	p, code := parseStatsPeriod(q, now, 30)
	if code != response.CodeSuccess {
		return nil, code
	}
	rows, err := s.statsDAO().GetTrendByCreatedDate(p.Start, p.End)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return buildTrendResponse(p, rows), response.CodeSuccess
}

func (s *AdminStatsServiceGroup) Locations(q *model.StatsDateQuery, limit int, now time.Time) (*model.LocationsResponse, response.Code) {
	if limit < 0 || limit > 100 {
		return nil, response.CodeParamError
	}
	p, code := parseStatsPeriod(q, now, 30)
	if code != response.CodeSuccess {
		return nil, code
	}
	rows, total, err := s.statsDAO().GetLocationsByCreatedPeriod(p.Start, p.End, limit)
	if errors.Is(err, dao.ErrStatsLocationLimit) {
		return nil, response.CodeParamError
	}
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return buildLocationsResponse(rows, total, limit), response.CodeSuccess
}

func normalizeStatsItemsQuery(q model.StatsItemsQuery, highView bool) (model.StatsItemsQuery, response.Code) {
	if q.Days == 0 {
		q.Days = 7
	}
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 10
	}
	if highView && q.MinViews == 0 {
		q.MinViews = 50
	}
	if q.Days < 1 || q.Days > 3650 || q.Page < 1 || q.Page > 10000 || q.PageSize < 1 || q.PageSize > 100 || (highView && (q.MinViews < 1 || q.MinViews > 2147483647)) {
		return q, response.CodeParamError
	}
	return q, response.CodeSuccess
}

func (s *AdminStatsServiceGroup) StatsItems(q model.StatsItemsQuery, now time.Time, highView bool) (*model.StatsItemsResponse, response.Code) {
	q, code := normalizeStatsItemsQuery(q, highView)
	if code != response.CodeSuccess {
		return nil, code
	}
	now = now.In(statsLocation)
	rows, total, err := s.statsDAO().GetStatsItems(q, now, now.AddDate(0, 0, -q.Days), highView)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return &model.StatsItemsResponse{Total: total, Page: q.Page, PageSize: q.PageSize, Items: rows}, response.CodeSuccess
}

func normalizeDurationQuery(q model.StatsDurationQuery, now time.Time) (model.StatsDurationQuery, model.StatsPeriod, response.Code) {
	if q.GroupBy == "" {
		q.GroupBy = "none"
	}
	if q.GroupBy != "none" && q.GroupBy != "location" && q.GroupBy != "type" {
		return q, model.StatsPeriod{}, response.CodeParamError
	}
	p, code := parseStatsPeriod(&q.StatsDateQuery, now, 30)
	return q, p, code
}

func (s *AdminStatsServiceGroup) ReturnDuration(q model.StatsDurationQuery, now time.Time) (*model.StatsDurationResponse, response.Code) {
	q, p, code := normalizeDurationQuery(q, now)
	if code != response.CodeSuccess {
		return nil, code
	}
	overall, groups, err := s.statsDAO().GetReturnDuration(p.Start, p.End, q.GroupBy)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return &model.StatsDurationResponse{StartDate: p.StartDate, EndDate: p.EndDate, GroupBy: q.GroupBy, Approximate: true, Overall: overall, Groups: groups}, response.CodeSuccess
}

func buildHeatmapResponse(p model.StatsPeriod, rows []model.StatsHeatmapRow) *model.HeatmapResponse {
	result := &model.HeatmapResponse{StartDate: p.StartDate, EndDate: p.EndDate}
	for _, row := range rows {
		if row.Weekday >= 0 && row.Weekday < 7 && row.Hour >= 0 && row.Hour < 24 {
			result.Matrix[row.Weekday][row.Hour] = row.Count
		}
	}
	return result
}

func (s *AdminStatsServiceGroup) TimeHeatmap(q *model.StatsDateQuery, now time.Time) (*model.HeatmapResponse, response.Code) {
	p, code := parseStatsPeriod(q, now, 30)
	if code != response.CodeSuccess {
		return nil, code
	}
	rows, err := s.statsDAO().GetTimeHeatmap(p.Start, p.End)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return buildHeatmapResponse(p, rows), response.CodeSuccess
}

func statsPercent(numerator, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	n := new(big.Int).Mul(big.NewInt(numerator), big.NewInt(10000))
	d := big.NewInt(denominator)
	q, r := new(big.Int), new(big.Int)
	q.QuoRem(n, d, r)
	if new(big.Int).Mul(r, big.NewInt(2)).Cmp(d) >= 0 {
		q.Add(q, big.NewInt(1))
	}
	v, _ := new(big.Rat).SetFrac(q, big.NewInt(100)).Float64()
	return v
}

func buildDistributionResponse(q model.StatsDistributionQuery, p model.StatsPeriod, rows []model.StatsDistributionRow, total int64) *model.StatsDistributionResponse {
	if rows == nil {
		rows = make([]model.StatsDistributionRow, 0)
	}
	for i := range rows {
		rows[i].Percent = statsPercent(rows[i].Published, total)
		rows[i].ReturnRate = statsPercent(rows[i].Returned, rows[i].Published)
	}
	return &model.StatsDistributionResponse{StartDate: p.StartDate, EndDate: p.EndDate, Dimension: q.Dimension, Total: total, Buckets: rows}
}

func (s *AdminStatsServiceGroup) Distribution(q model.StatsDistributionQuery, now time.Time) (*model.StatsDistributionResponse, response.Code) {
	if q.Dimension != "type" && q.Dimension != "tag" {
		return nil, response.CodeParamError
	}
	p, code := parseStatsPeriod(&q.StatsDateQuery, now, 30)
	if code != response.CodeSuccess {
		return nil, code
	}
	rows, total, err := s.statsDAO().GetDistribution(p.Start, p.End, q.Dimension)
	if errors.Is(err, dao.ErrStatsTagLimit) {
		return nil, response.CodeParamError
	}
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return buildDistributionResponse(q, p, rows, total), response.CodeSuccess
}

func buildFunnelResponse(p model.StatsPeriod, rows []model.StatsFunnelRow) *model.StatsFunnelResponse {
	ratios := make([]float64, 0, 3)
	for i := 1; i < len(rows); i++ {
		ratios = append(ratios, statsPercent(rows[i].Users, rows[i-1].Users))
	}
	return &model.StatsFunnelResponse{StartDate: p.StartDate, EndDate: p.EndDate, Stages: rows, Ratios: ratios}
}

func (s *AdminStatsServiceGroup) Funnel(q model.StatsDateQuery, now time.Time) (*model.StatsFunnelResponse, response.Code) {
	p, code := parseStatsPeriod(&q, now, 30)
	if code != response.CodeSuccess {
		return nil, code
	}
	rows, err := s.statsDAO().GetFunnel(p.Start, p.End)
	if err != nil {
		return nil, response.CodeDatabaseError
	}
	return buildFunnelResponse(p, rows), response.CodeSuccess
}
