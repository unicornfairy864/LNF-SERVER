package advanced

import (
	"encoding/json"
	"testing"
	"time"

	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"github.com/unicornfairy864/LNF-SERVER/response"
)

func TestParseStatsPeriod(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, statsLocation)
	tests := []struct {
		name       string
		q          model.StatsDateQuery
		wantCode   response.Code
		start, end string
	}{
		{"default", model.StatsDateQuery{}, response.CodeSuccess, "2026-09-04", "2026-10-03"},
		{"explicit", model.StatsDateQuery{StartDate: "2026-09-01", EndDate: "2026-09-03"}, response.CodeSuccess, "2026-09-01", "2026-09-03"},
		{"mixed", model.StatsDateQuery{Days: 3, StartDate: "2026-10-01", EndDate: "2026-10-03"}, response.CodeParamError, "", ""},
		{"future", model.StatsDateQuery{StartDate: "2026-10-04", EndDate: "2026-10-05"}, response.CodeParamError, "", ""},
		{"single day", model.StatsDateQuery{Days: 1}, response.CodeSuccess, "2026-10-03", "2026-10-03"},
		{"maximum", model.StatsDateQuery{StartDate: "2025-10-03", EndDate: "2026-10-03"}, response.CodeSuccess, "2025-10-03", "2026-10-03"},
		{"too long", model.StatsDateQuery{StartDate: "2025-10-02", EndDate: "2026-10-03"}, response.CodeParamError, "", ""},
		{"invalid date", model.StatsDateQuery{StartDate: "2026-02-29", EndDate: "2026-03-01"}, response.CodeParamError, "", ""},
		{"missing end", model.StatsDateQuery{StartDate: "2026-10-01"}, response.CodeParamError, "", ""},
		{"reverse", model.StatsDateQuery{StartDate: "2026-10-03", EndDate: "2026-10-01"}, response.CodeParamError, "", ""},
		{"negative days", model.StatsDateQuery{Days: -1}, response.CodeParamError, "", ""},
		{"too many days", model.StatsDateQuery{Days: 367}, response.CodeParamError, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, code := parseStatsPeriod(&tt.q, now, 30)
			if code != tt.wantCode || got.StartDate != tt.start || got.EndDate != tt.end {
				t.Fatalf("got period=%+v code=%v", got, code)
			}
		})
	}
}

func TestOverviewPeriods(t *testing.T) {
	tests := []struct {
		name                                   string
		now                                    time.Time
		q                                      model.StatsDateQuery
		start, end, previousStart, previousEnd string
	}{
		{"partial month", time.Date(2026, 10, 3, 12, 0, 0, 0, statsLocation), model.StatsDateQuery{}, "2026-10-01", "2026-10-03", "2026-09-01", "2026-09-30"},
		{"year boundary", time.Date(2026, 1, 1, 12, 0, 0, 0, statsLocation), model.StatsDateQuery{}, "2026-01-01", "2026-01-01", "2025-12-01", "2025-12-31"},
		{"leap month", time.Date(2024, 3, 31, 12, 0, 0, 0, statsLocation), model.StatsDateQuery{}, "2024-03-01", "2024-03-31", "2024-02-01", "2024-02-29"},
		{"explicit", time.Date(2024, 3, 2, 12, 0, 0, 0, statsLocation), model.StatsDateQuery{StartDate: "2024-02-28", EndDate: "2024-03-01"}, "2024-02-28", "2024-03-01", "2024-02-25", "2024-02-27"},
		{"UTC crosses day", time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC), model.StatsDateQuery{}, "2026-10-01", "2026-10-01", "2026-09-01", "2026-09-30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current, previous, code := overviewPeriods(&tt.q, tt.now)
			if code != response.CodeSuccess || current.StartDate != tt.start || current.EndDate != tt.end || previous.StartDate != tt.previousStart || previous.EndDate != tt.previousEnd || !previous.End.Equal(current.Start) {
				t.Fatalf("current=%+v previous=%+v code=%v", current, previous, code)
			}
		})
	}
}

func TestTrendZeroFilling(t *testing.T) {
	q := model.StatsDateQuery{StartDate: "2024-02-28", EndDate: "2024-03-01"}
	p, _ := parseStatsPeriod(&q, time.Date(2024, 3, 2, 0, 0, 0, 0, statsLocation), 30)
	for _, rows := range [][]model.StatsTrendRow{nil, {{Date: p.Start.AddDate(0, 0, 1), Published: 5, Returned: 2}}} {
		got := buildTrendResponse(p, rows)
		if len(got.Points) != 3 || got.Points[0].Date != "2024-02-28" || got.Points[1].Date != "2024-02-29" || got.Points[2].Date != "2024-03-01" || got.Points[0].Published != 0 || got.Points[2].Returned != 0 {
			t.Fatalf("unexpected points: %+v", got.Points)
		}
		if len(rows) > 0 && (got.Points[1].Published != 5 || got.Points[1].Returned != 2) {
			t.Fatal("lost populated day")
		}
	}
}

func TestRateMetricSerialization(t *testing.T) {
	for _, tt := range []struct {
		value, previous float64
		comparable      bool
		change          *float64
	}{
		{40, 25, true, floatPtr(60)}, {0, 0, true, floatPtr(0)}, {40, 0, false, nil}, {0, 25, true, floatPtr(-100)},
	} {
		got := rateMetric(tt.value, tt.previous)
		if got.Comparable != tt.comparable || (got.ChangePercent == nil) != (tt.change == nil) || (tt.change != nil && *got.ChangePercent != *tt.change) {
			t.Fatalf("unexpected metric: %+v", got)
		}
		data, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		var decoded model.StatsRateMetric
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Value != tt.value || decoded.Previous != tt.previous {
			t.Fatalf("invalid JSON scale: %s", data)
		}
	}
}

func TestLocationsLimiting(t *testing.T) {
	a, b := int64(1), int64(2)
	rows := []model.StatsLocationRow{{LocationID: &a, Name: "A", Count: 1}, {Count: 1}, {LocationID: &b, Name: "B", Count: 1}}
	for _, limit := range []int{0, 1, 4} {
		got := buildLocationsResponse(rows, 3, limit)
		want := 2
		if limit == 1 {
			want = 1
		}
		if len(got.Locations) != want || got.Total != 3 || got.Unknown.Count != 1 || got.Unknown.Percent != 33.33 || got.Locations[0].Percent != 33.33 || *got.Locations[0].LocationID != a {
			t.Fatalf("limit=%d response=%+v", limit, got)
		}
	}
	got := buildLocationsResponse(nil, 0, 0)
	if got.Locations == nil || len(got.Locations) != 0 || got.Unknown.Percent != 0 {
		t.Fatalf("empty response=%+v", got)
	}
}

func TestMetricZeroPrevious(t *testing.T) {
	got := metric(2, 0)
	if got.Comparable || got.ChangePercent != nil {
		t.Fatalf("unexpected metric: %+v", got)
	}
	got = metric(0, 0)
	if !got.Comparable || got.ChangePercent == nil || *got.ChangePercent != 0 {
		t.Fatalf("unexpected zero metric: %+v", got)
	}
}

func TestHeatmapMatrix(t *testing.T) {
	for _, rows := range [][]model.StatsHeatmapRow{nil, {{Weekday: 0, Hour: 0, Count: 2}, {Weekday: 6, Hour: 23, Count: 3}}} {
		got := buildHeatmapResponse(model.StatsPeriod{StartDate: "2026-10-01", EndDate: "2026-10-03"}, rows)
		var total int64
		for _, day := range got.Matrix {
			for _, count := range day {
				total += count
			}
		}
		if len(rows) == 0 && total != 0 {
			t.Fatal("empty matrix not zero")
		}
		if len(rows) > 0 && (total != 5 || got.Matrix[0][0] != 2 || got.Matrix[6][23] != 3) {
			t.Fatalf("matrix=%v", got.Matrix)
		}
		if got.StartDate != "2026-10-01" || got.EndDate != "2026-10-03" {
			t.Fatal("missing dates")
		}
	}
}

func TestStatsItemsQuery(t *testing.T) {
	for _, tt := range []struct {
		q     model.StatsItemsQuery
		high  bool
		valid bool
	}{
		{model.StatsItemsQuery{}, false, true}, {model.StatsItemsQuery{}, true, true},
		{model.StatsItemsQuery{Days: 3650, Page: 10000, PageSize: 100, MinViews: 2147483647}, true, true},
		{model.StatsItemsQuery{Days: -1}, false, false}, {model.StatsItemsQuery{Days: 3651}, false, false},
		{model.StatsItemsQuery{Page: -1}, false, false}, {model.StatsItemsQuery{Page: 10001}, false, false},
		{model.StatsItemsQuery{PageSize: 101}, false, false}, {model.StatsItemsQuery{PageSize: -1}, false, false},
		{model.StatsItemsQuery{MinViews: -1}, true, false}, {model.StatsItemsQuery{MinViews: 2147483648}, true, false},
	} {
		got, code := normalizeStatsItemsQuery(tt.q, tt.high)
		if (code == response.CodeSuccess) != tt.valid {
			t.Fatalf("query=%+v code=%v", tt.q, code)
		}
		if tt.valid && tt.q == (model.StatsItemsQuery{}) && (got.Days != 7 || got.Page != 1 || got.PageSize != 10 || (tt.high && got.MinViews != 50)) {
			t.Fatalf("defaults=%+v", got)
		}
	}
}

func TestDurationQuery(t *testing.T) {
	now := time.Date(2026, 10, 3, 12, 0, 0, 0, statsLocation)
	for _, dimension := range []string{"", "none", "type", "location", "contact", "type; DROP TABLE items"} {
		q, p, code := normalizeDurationQuery(model.StatsDurationQuery{GroupBy: dimension}, now)
		valid := dimension == "" || dimension == "none" || dimension == "type" || dimension == "location"
		if (code == response.CodeSuccess) != valid {
			t.Fatalf("dimension=%s code=%v", dimension, code)
		}
		if valid && (p.StartDate != "2026-09-04" || q.GroupBy == "") {
			t.Fatalf("query=%+v period=%+v", q, p)
		}
	}
}

func TestStatsPercent(t *testing.T) {
	for _, tt := range []struct {
		n, d int64
		want float64
	}{{0, 0, 0}, {1, 3, 33.33}, {2, 3, 66.67}, {1, 32, 3.13}, {3, 2, 150}, {9223372036854775807, 9223372036854775807, 100}} {
		if got := statsPercent(tt.n, tt.d); got != tt.want {
			t.Fatalf("%d/%d: %v", tt.n, tt.d, got)
		}
	}
}

func TestDistributionResponse(t *testing.T) {
	id := int64(1)
	got := buildDistributionResponse(model.StatsDistributionQuery{Dimension: "tag"}, model.StatsPeriod{}, []model.StatsDistributionRow{{BucketID: &id, Published: 2, Returned: 1}, {Published: 1}}, 2)
	if got.Buckets[0].Percent != 100 || got.Buckets[0].ReturnRate != 50 || got.Buckets[1].Percent != 50 {
		t.Fatal(got)
	}
	empty := buildDistributionResponse(model.StatsDistributionQuery{Dimension: "type"}, model.StatsPeriod{}, nil, 0)
	if empty.Buckets == nil {
		t.Fatal("empty buckets must be array")
	}
	for _, dimension := range []string{"", "contact", "tag;DROP TABLE items"} {
		_, code := AdminStatsService.Distribution(model.StatsDistributionQuery{Dimension: dimension}, time.Now())
		if code != response.CodeParamError {
			t.Fatal(dimension)
		}
	}
}

func TestFunnelRatios(t *testing.T) {
	got := buildFunnelResponse(model.StatsPeriod{}, []model.StatsFunnelRow{{Users: 0}, {Users: 2}, {Users: 3}, {Users: 1}})
	for i, want := range []float64{0, 150, 33.33} {
		if got.Ratios[i] != want {
			t.Fatal(got)
		}
	}
}
