package followup

import (
	"testing"
	"time"
)

func TestScheduleKeepsAStableDeliveryKey(t *testing.T) {
	now := time.Date(2026, 7, 31, 9, 0, 0, 0, time.UTC)
	first := Schedule("ord_1042", "cus_77", 3, now)
	second := Schedule("ord_1042", "cus_77", 3, now)
	if !first.DueAt.Equal(now.Add(3 * time.Hour)) {
		t.Fatalf("due time = %s", first.DueAt)
	}
	if first.DeliveryKey != second.DeliveryKey {
		t.Fatal("expected a stable delivery key")
	}
}
