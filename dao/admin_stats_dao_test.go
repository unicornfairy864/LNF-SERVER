package dao

import (
	model "github.com/unicornfairy864/LNF-SERVER/model/advanced"
	"strings"
	"testing"
)

func TestStatsLocationLimit(t *testing.T) {
	for _, tt := range []struct {
		count, limit      int
		unknown, exceeded bool
	}{
		{0, 0, false, false}, {1000, 0, true, false}, {1001, 0, false, true}, {1001, 0, true, true}, {1001, 4, true, false},
	} {
		rows := make([]model.StatsLocationRow, tt.count)
		for i := range rows {
			id := int64(i + 1)
			rows[i].LocationID = &id
		}
		if tt.unknown {
			rows = append(rows, model.StatsLocationRow{})
		}
		if got := statsLocationLimitExceeded(rows, tt.limit); got != tt.exceeded {
			t.Fatalf("count=%d limit=%d unknown=%v: got %v", tt.count, tt.limit, tt.unknown, got)
		}
	}
}

func TestStatsDurationSQLWhitelist(t *testing.T) {
	for _, dimension := range []string{"none", "location", "type", "", "location_id; DROP TABLE items"} {
		query, ok := statsDurationSQL(dimension)
		valid := dimension == "none" || dimension == "location" || dimension == "type"
		if ok != valid {
			t.Fatalf("dimension=%s allowed=%v", dimension, ok)
		}
		if valid {
			for _, required := range []string{"is_deleted=0", "claim_time IS NOT NULL", "updated_at >= ? AND updated_at < ?", "updated_at >= created_at", "ROW_NUMBER()", "FLOOR((n+1)/2)", "FLOOR((n+2)/2)"} {
				if !strings.Contains(query, required) {
					t.Fatalf("missing %s", required)
				}
			}
		}
	}
}

func TestStatsDistributionSQL(t *testing.T) {
	for _, dimension := range []string{"type", "tag", "", "contact", "tag;DROP TABLE items"} {
		query, ok := statsDistributionSQL(dimension)
		if ok != (dimension == "type" || dimension == "tag") {
			t.Fatal(dimension)
		}
		if ok {
			for _, required := range []string{"i.is_deleted=0", "i.created_at >= ? AND i.created_at < ?", "i.status=2", "i.claim_user_id IS NOT NULL", "i.claim_time IS NOT NULL", "ORDER BY"} {
				if !strings.Contains(query, required) {
					t.Fatal(required)
				}
			}
		}
		if dimension == "tag" && (!strings.Contains(query, "COUNT(DISTINCT i.id)") || !strings.Contains(query, "LIMIT 1002")) {
			t.Fatal("missing distinct or cap")
		}
	}
}

func TestStatsFunnelSQL(t *testing.T) {
	if strings.Count(statsFunnelSQL, "?") != 14 {
		t.Fatal("invalid placeholders")
	}
	for _, required := range []string{"users WHERE is_deleted=0", "FROM items WHERE is_deleted=0", "claim_time >= ? AND claim_time < ?", "status=2 AND claim_time IS NOT NULL", "updated_at >= ? AND updated_at < ?", "COUNT(DISTINCT"} {
		if !strings.Contains(statsFunnelSQL, required) {
			t.Fatal(required)
		}
	}
}

func TestStatsTagLimit(t *testing.T) {
	for _, count := range []int{0, 1000, 1001} {
		rows := make([]model.StatsDistributionRow, count)
		for i := range rows {
			id := int64(i + 1)
			rows[i].BucketID = &id
		}
		rows = append(rows, model.StatsDistributionRow{BucketName: "untagged"})
		if statsTagLimitExceeded(rows) != (count > 1000) {
			t.Fatal(count)
		}
	}
}
