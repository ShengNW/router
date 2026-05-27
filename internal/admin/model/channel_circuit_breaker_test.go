package model

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newChannelCircuitBreakerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=private"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&ChannelCircuitBreakerState{}); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	return db
}

func TestRecordChannelCircuitBreakerOpenAndRecovered(t *testing.T) {
	db := newChannelCircuitBreakerTestDB(t)

	if err := recordChannelCircuitBreakerOpenWithDB(db, "channel-1", "low_success_rate", 0.25, 12345); err != nil {
		t.Fatalf("record open: %v", err)
	}
	row, err := getChannelCircuitBreakerStateWithDB(db, "channel-1")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if row.State != ChannelCircuitBreakerStateOpen || row.Reason != "low_success_rate" || row.SuccessRate != 0.25 || row.RecoverAfter != 12345 {
		t.Fatalf("open state = %+v, want low success rate open row", row)
	}

	if err := updateChannelCircuitBreakerStateWithDB(db, "channel-1", ChannelCircuitBreakerStateRecovered, ""); err != nil {
		t.Fatalf("record recovered: %v", err)
	}
	row, err = getChannelCircuitBreakerStateWithDB(db, "channel-1")
	if err != nil {
		t.Fatalf("get recovered state: %v", err)
	}
	if row.State != ChannelCircuitBreakerStateRecovered || row.RecoveredAt == 0 {
		t.Fatalf("recovered state = %+v, want recovered with recovered_at", row)
	}
}

func TestRecordChannelCircuitBreakerCanceledOnlyUpdatesOpenState(t *testing.T) {
	db := newChannelCircuitBreakerTestDB(t)

	if err := recordChannelCircuitBreakerOpenWithDB(db, "channel-1", "low_success_rate", 0.25, 12345); err != nil {
		t.Fatalf("record open: %v", err)
	}
	if err := updateChannelCircuitBreakerStateWithDB(db, "channel-1", ChannelCircuitBreakerStateCanceled, "insufficient balance"); err != nil {
		t.Fatalf("record canceled: %v", err)
	}
	if err := updateChannelCircuitBreakerStateWithDB(db, "channel-1", ChannelCircuitBreakerStateRecovered, ""); err != nil {
		t.Fatalf("record recovered after canceled: %v", err)
	}
	row, err := getChannelCircuitBreakerStateWithDB(db, "channel-1")
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if row.State != ChannelCircuitBreakerStateCanceled || row.Reason != "insufficient balance" {
		t.Fatalf("state = %+v, want canceled and not overwritten by recovery", row)
	}
}

func TestListOpenChannelCircuitBreakerStates(t *testing.T) {
	db := newChannelCircuitBreakerTestDB(t)

	if err := recordChannelCircuitBreakerOpenWithDB(db, "channel-1", "low_success_rate", 0.25, 12345); err != nil {
		t.Fatalf("record open: %v", err)
	}
	if err := recordChannelCircuitBreakerOpenWithDB(db, "channel-2", "low_success_rate", 0.75, 12346); err != nil {
		t.Fatalf("record open: %v", err)
	}
	if err := updateChannelCircuitBreakerStateWithDB(db, "channel-2", ChannelCircuitBreakerStateRecovered, ""); err != nil {
		t.Fatalf("record recovered: %v", err)
	}

	rows, err := listOpenChannelCircuitBreakerStatesWithDB(db)
	if err != nil {
		t.Fatalf("list open: %v", err)
	}
	if len(rows) != 1 || rows[0].ChannelId != "channel-1" {
		t.Fatalf("open rows = %+v, want channel-1 only", rows)
	}
}
