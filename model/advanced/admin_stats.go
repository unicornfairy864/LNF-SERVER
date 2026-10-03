package model

import "time"

// StatsDateQuery 是统计接口通用日期参数。
type StatsDateQuery struct {
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Days      int    `form:"days"`
}

// StatsPeriod 是经过 service 校验后的东八区半开时间范围。
type StatsPeriod struct {
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	Start     time.Time `json:"-"`
	End       time.Time `json:"-"`
}

type StatsMetric struct {
	Value         int64    `json:"value"`
	Previous      int64    `json:"previous"`
	ChangePercent *float64 `json:"change_percent"`
	Comparable    bool     `json:"comparable"`
}

type StatsRateMetric struct {
	Value         float64  `json:"value"`
	Previous      float64  `json:"previous"`
	ChangePercent *float64 `json:"change_percent"`
	Comparable    bool     `json:"comparable"`
}

type OverviewResponse struct {
	Period         StatsPeriod     `json:"period"`
	Published      StatsMetric     `json:"published"`
	Returned       StatsMetric     `json:"returned"`
	Pending        StatsMetric     `json:"pending"`
	ReturnRate     StatsRateMetric `json:"return_rate"`
	PendingOver24h int64           `json:"pending_over_24h"`
}

type TrendPoint struct {
	Date      string `json:"date"`
	Published int64  `json:"published"`
	Returned  int64  `json:"returned"`
}

type TrendResponse struct {
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
	Points    []TrendPoint `json:"points"`
}

type LocationStat struct {
	LocationID *int64  `json:"location_id,omitempty"`
	Name       string  `json:"name"`
	Count      int64   `json:"count"`
	Percent    float64 `json:"percent"`
}

type LocationsResponse struct {
	Total     int64          `json:"total"`
	Locations []LocationStat `json:"locations"`
	Unknown   LocationStat   `json:"unknown"`
}

// StatsOverviewRow 是 DAO 返回的单个发布批次状态聚合。
type StatsOverviewRow struct {
	Published int64
	Returned  int64
	Pending   int64
	Claimed   int64
	Closed    int64
}

type StatsTrendRow struct {
	Date      time.Time
	Published int64
	Returned  int64
}

type StatsLocationRow struct {
	LocationID *int64
	Name       string
	Count      int64
}

type StatsHeatmapRow struct {
	Weekday int
	Hour    int
	Count   int64
}

type HeatmapResponse struct {
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
	Matrix    [7][24]int64 `json:"matrix"`
}

type StatsItemsQuery struct {
	Days     int   `form:"days"`
	Page     int   `form:"page"`
	PageSize int   `form:"page_size"`
	MinViews int64 `form:"min_views"`
}

type StatsItemRow struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	Type         int8      `json:"type"`
	Status       int8      `json:"status"`
	LocationID   *int64    `json:"location_id"`
	LocationName string    `json:"location_name"`
	ViewCount    int64     `json:"view_count"`
	CreatedAt    time.Time `json:"created_at"`
	StagnantDays int64     `json:"stagnant_days"`
}

type StatsItemsResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Items    []StatsItemRow `json:"items"`
}

type StatsDurationQuery struct {
	StatsDateQuery
	GroupBy string `form:"group_by"`
}

type StatsDurationRow struct {
	GroupID        *int64  `json:"group_id"`
	Count          int64   `json:"count"`
	AverageSeconds float64 `json:"average_seconds"`
	MedianSeconds  float64 `json:"median_seconds"`
}

type StatsDurationResponse struct {
	StartDate   string             `json:"start_date"`
	EndDate     string             `json:"end_date"`
	GroupBy     string             `json:"group_by"`
	Approximate bool               `json:"approximate"`
	Overall     StatsDurationRow   `json:"overall"`
	Groups      []StatsDurationRow `json:"groups"`
}

type StatsDistributionQuery struct {
	StatsDateQuery
	Dimension string `form:"dimension"`
}

type StatsDistributionRow struct {
	BucketID   *int64  `json:"bucket_id"`
	BucketName string  `json:"bucket_name"`
	Published  int64   `json:"published"`
	Returned   int64   `json:"returned"`
	ReturnRate float64 `json:"return_rate"`
	Percent    float64 `json:"percent"`
}

type StatsDistributionResponse struct {
	StartDate string                 `json:"start_date"`
	EndDate   string                 `json:"end_date"`
	Dimension string                 `json:"dimension"`
	Total     int64                  `json:"total"`
	Buckets   []StatsDistributionRow `json:"buckets"`
}

type StatsFunnelRow struct {
	Stage string `json:"stage"`
	Users int64  `json:"users"`
}

type StatsFunnelResponse struct {
	StartDate string           `json:"start_date"`
	EndDate   string           `json:"end_date"`
	Stages    []StatsFunnelRow `json:"stages"`
	Ratios    []float64        `json:"adjacent_ratios"`
}
