package test

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
)

func TestStatsCSV(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rows := make([][]string, 401)
	for i := range rows {
		rows[i] = []string{"教学楼,\"A\"\n一层", "33.33"}
	}
	if err := writeStatsCSV(ctx, c, "locations", "2026-10-01", "2026-10-03", []string{"name", "percent"}, rows); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rec.Body.String(), "\xef\xbb\xbf") || rec.Header().Get("Content-Type") != "text/csv; charset=utf-8" || !rec.Flushed {
		t.Fatal("missing CSV headers, BOM or flush")
	}
	want := url.PathEscape("管理员统计-locations-2026-10-01-2026-10-03.csv")
	if !strings.Contains(rec.Header().Get("Content-Disposition"), "filename*=UTF-8''"+want) {
		t.Fatal("invalid filename")
	}
	parsed, err := csv.NewReader(strings.NewReader(strings.TrimPrefix(rec.Body.String(), "\xef\xbb\xbf"))).ReadAll()
	if err != nil || len(parsed) != 402 || parsed[1][0] != rows[0][0] {
		t.Fatalf("CSV escaping: %v", err)
	}
}

func TestStatsCSVBeforeWriteErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		cancel bool
		count  int
	}{{"cancelled", true, 1}, {"too many", false, 10001}} {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel()
			}
			if err := writeStatsCSV(ctx, c, "trend", "a", "b", []string{"date"}, make([][]string, tt.count)); err == nil || c.Writer.Written() {
				t.Fatal("failure must occur before writing")
			}
		})
	}
}

type statsFailWriter struct{ *httptest.ResponseRecorder }

func (w statsFailWriter) Write(p []byte) (int, error) { return 0, errors.New("disconnected") }

func TestStatsCSVWriterFailure(t *testing.T) {
	c, _ := gin.CreateTestContext(statsFailWriter{httptest.NewRecorder()})
	if err := writeStatsCSV(context.Background(), c, "trend", "a", "b", []string{"date"}, nil); err == nil {
		t.Fatal("write error ignored")
	}
}

func TestStatsInvalidParametersJSON(t *testing.T) {
	h := &AdminStatsHandlerGroup{}
	for _, tt := range []struct {
		path    string
		handler gin.HandlerFunc
	}{
		{"/?export=xml", h.Trend}, {"/?export=csv&days=-1", h.Trend}, {"/?export=csv&limit=bad", h.Locations},
		{"/?dimension=contact", h.Distribution}, {"/?days=367", h.Funnel},
	} {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest("GET", tt.path, nil)
		tt.handler(c)
		var envelope struct {
			Code int                    `json:"code"`
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil || rec.Code != 200 || envelope.Code != 1 || envelope.Data == nil || len(envelope.Data) != 0 {
			t.Fatalf("%s: %s", tt.path, rec.Body.String())
		}
	}
}

func TestStatsCSVReports(t *testing.T) {
	for _, data := range []interface{}{
		&model.OverviewResponse{}, &model.TrendResponse{}, &model.LocationsResponse{},
		&model.StatsDistributionResponse{}, &model.StatsDurationResponse{}, &model.HeatmapResponse{},
	} {
		header, rows, _, _ := statsCSVRows(data)
		if len(header) == 0 {
			t.Fatalf("missing header for %T", data)
		}
		for _, row := range rows {
			if len(row) != len(header) {
				t.Fatalf("%T inconsistent columns", data)
			}
		}
	}
}
