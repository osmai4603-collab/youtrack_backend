package sqlstore

import (
	"context"
	"testing"
)

func TestSqlStoreReadyFailsWithoutDatabase(t *testing.T) {
	s := &SqlStore{}
	if err := s.Ready(context.Background()); err == nil {
		t.Fatal("expected database readiness check to fail when store is not connected")
	}
}
