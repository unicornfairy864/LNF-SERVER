package advanced

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
	serviceAdvanced "github.com/unicornfairy864/LNF-SERVER/service/advanced"
)

type AdminStatsHandlerGroup struct{}

// Overview 查询发布队列的当前状态概览。
// @Summary 管理员统计概览
// @Description role>=1；按 created_at 选取队列并判断当前状态。默认当月至今天，对比上个完整月；显式范围对比紧邻等长周期。日期与 days 互斥，最大366天。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param start_date query string false "开始日期 YYYY-MM-DD，须与 end_date 同传"
// @Param end_date query string false "结束日期 YYYY-MM-DD，包含当天，不晚于今天"
// @Param days query int false "最近自然日数，1-366，与日期对互斥"
// @Success 200 {object} response.CommonResponse{data=model.OverviewResponse}
// @Param export query string false "csv；最多10000数据行，30秒超时；成功text/csv含BOM，失败HTTP200 JSON，流中失败终止"
// @Produce text/csv
// @Router /api/v1/admin/stats/overview [get]
func (h *AdminStatsHandlerGroup) Overview(c *gin.Context) {
	if handleStatsCSV(c, "overview") {
		return
	}
	var q model.StatsDateQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.Overview(&q, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

// Trend 查询创建队列的自然日趋势。
// @Summary 管理员发布队列趋势
// @Description role>=1；published 与 returned 均归入物品创建日，returned 取当前成功归还快照；东八区自然日补零。默认30天，最大366天。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param start_date query string false "开始日期 YYYY-MM-DD，须与 end_date 同传"
// @Param end_date query string false "结束日期 YYYY-MM-DD，包含当天，不晚于今天"
// @Param days query int false "最近自然日数，默认30，1-366，与日期对互斥"
// @Success 200 {object} response.CommonResponse{data=model.TrendResponse}
// @Param export query string false "csv；最多10000数据行，30秒超时；成功text/csv含BOM，失败HTTP200 JSON，流中失败终止"
// @Produce text/csv
// @Router /api/v1/admin/stats/trend [get]
func (h *AdminStatsHandlerGroup) Trend(c *gin.Context) {
	if handleStatsCSV(c, "trend") {
		return
	}
	var q model.StatsDateQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.Trend(&q, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

// Locations 查询直接地点发布量。
// @Summary 管理员地点统计
// @Description role>=1；按直接 location_id 聚合，NULL 单列 unknown，不向父节点归并。percent 分母为期内全部未删除发布量。排序 count DESC、location_id ASC。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param start_date query string false "开始日期 YYYY-MM-DD，须与 end_date 同传"
// @Param end_date query string false "结束日期 YYYY-MM-DD，包含当天，不晚于今天"
// @Param days query int false "最近自然日数，默认30，1-366，与日期对互斥"
// @Param limit query int false "地点桶数量，默认4，0-100；0最多返回1000个地点桶，超限code=1；unknown 单独返回"
// @Success 200 {object} response.CommonResponse{data=model.LocationsResponse}
// @Param export query string false "csv；最多10000数据行，30秒超时；保留limit参数，失败HTTP200 JSON，流中失败终止"
// @Produce text/csv
// @Router /api/v1/admin/stats/locations [get]
func (h *AdminStatsHandlerGroup) Locations(c *gin.Context) {
	if handleStatsCSV(c, "locations") {
		return
	}
	var q model.StatsDateQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	limit := 4
	if raw := c.Query("limit"); raw != "" {
		var err error
		limit, err = strconv.Atoi(raw)
		if err != nil {
			response.FailWithCode(c, response.CodeParamError)
			return
		}
	}
	data, code := serviceAdvanced.AdminStatsService.Locations(&q, limit, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

func nowForStats() time.Time { return time.Now() }

// Stagnant 查询滞留待处理清单。
// @Summary 管理员滞留待处理清单
// @Description role>=1；status=0且未删除，created_at<=当前时刻减days天；created_at ASC,id ASC。仅返回物品摘要，无联系方式。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param days query int false "滞留天数，默认7，1-3650"
// @Param page query int false "页码，默认1，1-10000"
// @Param page_size query int false "每页数量，默认10，1-100"
// @Success 200 {object} response.CommonResponse{data=model.StatsItemsResponse}
// @Router /api/v1/admin/stats/items/stagnant [get]
func (h *AdminStatsHandlerGroup) Stagnant(c *gin.Context) { h.statsItems(c, false) }

// HighView 查询高浏览待认领清单。
// @Summary 管理员高浏览低认领清单
// @Description role>=1；未删除且status=0，达到浏览与滞留阈值。排序view_count DESC,created_at ASC,id ASC。仅返回批准摘要字段。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param days query int false "滞留天数，默认7，1-3650"
// @Param min_views query int false "浏览下限，默认50，1-2147483647"
// @Param page query int false "页码，默认1，1-10000"
// @Param page_size query int false "每页数量，默认10，1-100"
// @Success 200 {object} response.CommonResponse{data=model.StatsItemsResponse}
// @Router /api/v1/admin/stats/items/high-view [get]
func (h *AdminStatsHandlerGroup) HighView(c *gin.Context) { h.statsItems(c, true) }

func (h *AdminStatsHandlerGroup) statsItems(c *gin.Context, highView bool) {
	var q model.StatsItemsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.StatsItems(q, nowForStats(), highView)
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

// ReturnDuration 查询近似归还时长。
// @Summary 管理员近似归还时长
// @Description role>=1；成功归还快照按updated_at纳入范围，时长updated_at-created_at，排除负时长。updated_at可被其他更新污染，结果为近似值。MySQL8窗口函数计算精确样本中位数；空样本各指标为0。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param start_date query string false "开始日期YYYY-MM-DD，与end_date同传"
// @Param end_date query string false "结束日期，包含当天，不晚于今天"
// @Param days query int false "最近自然日数，默认30，1-366，与日期对互斥"
// @Param group_by query string false "分组none|location|type，默认none"
// @Success 200 {object} response.CommonResponse{data=model.StatsDurationResponse}
// @Param export query string false "csv；最多10000数据行（含overall），30秒超时；失败HTTP200 JSON，流中失败终止"
// @Produce text/csv
// @Router /api/v1/admin/stats/return-duration [get]
func (h *AdminStatsHandlerGroup) ReturnDuration(c *gin.Context) {
	if handleStatsCSV(c, "return-duration") {
		return
	}
	var q model.StatsDurationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.ReturnDuration(q, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

// TimeHeatmap 查询东八区发布时段矩阵。
// @Summary 管理员发布时段热力图
// @Description role>=1；仅统计未删除物品，按created_at归属。matrix行是周一至周日，列是0-23点，空桶补零。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param start_date query string false "开始日期YYYY-MM-DD，与end_date同传"
// @Param end_date query string false "结束日期，包含当天，不晚于今天"
// @Param days query int false "最近自然日数，默认30，1-366，与日期对互斥"
// @Success 200 {object} response.CommonResponse{data=model.HeatmapResponse}
// @Param export query string false "csv；168数据行，30秒超时；失败HTTP200 JSON，流中失败终止"
// @Produce text/csv
// @Router /api/v1/admin/stats/time-heatmap [get]
func (h *AdminStatsHandlerGroup) TimeHeatmap(c *gin.Context) {
	if handleStatsCSV(c, "time-heatmap") {
		return
	}
	var q model.StatsDateQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.TimeHeatmap(&q, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

// Distribution queries current return snapshots by type or tag.
// @Summary 管理员维度统计
// @Description role>=1；默认30天、最大366天；tag可重复计数，percent分母为发布总数。最多1000标签桶，untagged单列；参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param dimension query string true "type|tag"
// @Param start_date query string false "开始日期YYYY-MM-DD"
// @Param end_date query string false "结束日期，与start_date同传"
// @Param days query int false "默认30，1-366，与日期对互斥"
// @Param export query string false "csv；最多10000行，超时30秒，成功为CSV流"
// @Success 200 {object} response.CommonResponse{data=model.StatsDistributionResponse}
// @Produce text/csv
// @Router /api/v1/admin/stats/distribution [get]
func (h *AdminStatsHandlerGroup) Distribution(c *gin.Context) {
	if handleStatsCSV(c, "distribution") {
		return
	}
	var q model.StatsDistributionQuery
	if c.ShouldBindQuery(&q) != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.Distribution(q, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

// Funnel queries independent activity layers, not a nested cohort.
// @Summary 管理员参与漏斗近似统计
// @Description role>=1；各层非嵌套，相邻比率可超过100%，分母0返回0。认领撤回历史无法恢复；归还按updated_at近似。参数错误code=1，数据库错误code=6。
// @Tags admin-stats
// @Produce json
// @Param Authorization header string true "Bearer JWT"
// @Param start_date query string false "开始日期YYYY-MM-DD"
// @Param end_date query string false "结束日期，与start_date同传"
// @Param days query int false "默认30，1-366，与日期对互斥"
// @Success 200 {object} response.CommonResponse{data=model.StatsFunnelResponse}
// @Router /api/v1/admin/stats/funnel [get]
func (h *AdminStatsHandlerGroup) Funnel(c *gin.Context) {
	var q model.StatsDateQuery
	if c.ShouldBindQuery(&q) != nil {
		response.FailWithCode(c, response.CodeParamError)
		return
	}
	data, code := serviceAdvanced.AdminStatsService.Funnel(q, nowForStats())
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return
	}
	response.SuccessWithData(c, data)
}

func handleStatsCSV(c *gin.Context, report string) bool {
	export := c.Query("export")
	if export == "" {
		return false
	}
	if export != "csv" {
		response.FailWithCode(c, response.CodeParamError)
		return true
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()
	var q model.StatsDistributionQuery
	if c.ShouldBindQuery(&q) != nil {
		response.FailWithCode(c, response.CodeParamError)
		return true
	}
	s := serviceAdvanced.AdminStatsServiceGroup{Context: ctx}
	var data interface{}
	var code response.Code
	switch report {
	case "overview":
		data, code = s.Overview(&q.StatsDateQuery, nowForStats())
	case "trend":
		data, code = s.Trend(&q.StatsDateQuery, nowForStats())
	case "locations":
		limit := 4
		if raw := c.Query("limit"); raw != "" {
			var err error
			limit, err = strconv.Atoi(raw)
			if err != nil {
				response.FailWithCode(c, response.CodeParamError)
				return true
			}
		}
		data, code = s.Locations(&q.StatsDateQuery, limit, nowForStats())
	case "distribution":
		data, code = s.Distribution(q, nowForStats())
	case "return-duration":
		var dq model.StatsDurationQuery
		if c.ShouldBindQuery(&dq) != nil {
			response.FailWithCode(c, response.CodeParamError)
			return true
		}
		data, code = s.ReturnDuration(dq, nowForStats())
	case "time-heatmap":
		data, code = s.TimeHeatmap(&q.StatsDateQuery, nowForStats())
	default:
		code = response.CodeParamError
	}
	if ctx.Err() != nil {
		code = response.CodeServerError
	}
	if code != response.CodeSuccess {
		response.FailWithCode(c, code)
		return true
	}
	header, rows, start, end := statsCSVRows(data)
	if report == "locations" {
		if q.StartDate != "" {
			start, end = q.StartDate, q.EndDate
		} else {
			loc := time.FixedZone("CST", 8*60*60)
			now := nowForStats().In(loc)
			days := q.Days
			if days == 0 {
				days = 30
			}
			end = now.Format("2006-01-02")
			start = now.AddDate(0, 0, 1-days).Format("2006-01-02")
		}
	}
	if len(rows) > 10000 {
		response.FailWithCode(c, response.CodeParamError)
		return true
	}
	if err := writeStatsCSV(ctx, c, report, start, end, header, rows); err != nil {
		if !c.Writer.Written() {
			response.FailWithCode(c, response.CodeServerError)
		}
		log.Printf("admin statistics CSV %s failed: %v", report, err)
	}
	return true
}

func statsCSVRows(data interface{}) ([]string, [][]string, string, string) {
	rows := make([][]string, 0)
	var header []string
	var start, end string
	i := func(v int64) string { return strconv.FormatInt(v, 10) }
	f := func(v float64) string { return strconv.FormatFloat(v, 'f', 2, 64) }
	id := func(v *int64) string {
		if v == nil {
			return ""
		}
		return i(*v)
	}
	switch d := data.(type) {
	case *model.OverviewResponse:
		start, end = d.Period.StartDate, d.Period.EndDate
		header = []string{"metric", "value", "previous", "change_percent", "comparable"}
		for _, m := range []struct {
			name   string
			metric model.StatsMetric
		}{{"published", d.Published}, {"returned", d.Returned}, {"pending", d.Pending}} {
			change := ""
			if m.metric.ChangePercent != nil {
				change = f(*m.metric.ChangePercent)
			}
			rows = append(rows, []string{m.name, i(m.metric.Value), i(m.metric.Previous), change, strconv.FormatBool(m.metric.Comparable)})
		}
		change := ""
		if d.ReturnRate.ChangePercent != nil {
			change = f(*d.ReturnRate.ChangePercent)
		}
		rows = append(rows, []string{"return_rate", f(d.ReturnRate.Value), f(d.ReturnRate.Previous), change, strconv.FormatBool(d.ReturnRate.Comparable)}, []string{"pending_over_24h", i(d.PendingOver24h), "", "", ""})
	case *model.TrendResponse:
		start, end = d.StartDate, d.EndDate
		header = []string{"date", "published", "returned"}
		for _, r := range d.Points {
			rows = append(rows, []string{r.Date, i(r.Published), i(r.Returned)})
		}
	case *model.LocationsResponse:
		header = []string{"location_id", "name", "count", "percent"}
		for _, r := range d.Locations {
			rows = append(rows, []string{id(r.LocationID), r.Name, i(r.Count), f(r.Percent)})
		}
		rows = append(rows, []string{"", d.Unknown.Name, i(d.Unknown.Count), f(d.Unknown.Percent)})
	case *model.StatsDistributionResponse:
		start, end = d.StartDate, d.EndDate
		header = []string{"bucket_id", "bucket_name", "published", "returned", "return_rate", "percent"}
		for _, r := range d.Buckets {
			rows = append(rows, []string{id(r.BucketID), r.BucketName, i(r.Published), i(r.Returned), f(r.ReturnRate), f(r.Percent)})
		}
	case *model.StatsDurationResponse:
		start, end = d.StartDate, d.EndDate
		header = []string{"scope", "group_id", "count", "average_seconds", "median_seconds", "approximate"}
		rows = append(rows, []string{"overall", "", i(d.Overall.Count), f(d.Overall.AverageSeconds), f(d.Overall.MedianSeconds), "true"})
		for _, r := range d.Groups {
			rows = append(rows, []string{d.GroupBy, id(r.GroupID), i(r.Count), f(r.AverageSeconds), f(r.MedianSeconds), "true"})
		}
	case *model.HeatmapResponse:
		start, end = d.StartDate, d.EndDate
		header = []string{"weekday", "hour", "published"}
		for day := 0; day < 7; day++ {
			for hour := 0; hour < 24; hour++ {
				rows = append(rows, []string{strconv.Itoa(day), strconv.Itoa(hour), i(d.Matrix[day][hour])})
			}
		}
	}
	return header, rows, start, end
}

func writeStatsCSV(ctx context.Context, c *gin.Context, report, start, end string, header []string, rows [][]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(rows) > 10000 {
		return errors.New("CSV row limit exceeded")
	}
	filename := fmt.Sprintf("管理员统计-%s-%s-%s.csv", report, start, end)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=stats.csv; filename*=UTF-8''"+url.PathEscape(filename))
	controller := http.NewResponseController(c.Writer)
	if deadline, ok := ctx.Deadline(); ok {
		_ = controller.SetWriteDeadline(deadline)
	}
	if _, err := c.Writer.Write([]byte{0xef, 0xbb, 0xbf}); err != nil {
		return err
	}
	w := csv.NewWriter(c.Writer)
	if err := w.Write(header); err != nil {
		return err
	}
	for j, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.Write(row); err != nil {
			return err
		}
		if (j+1)%200 == 0 {
			w.Flush()
			if err := w.Error(); err != nil {
				return err
			}
			c.Writer.Flush()
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}
