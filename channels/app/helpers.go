package app

import (
	"time"

	"github.com/google/uuid"
)

// newIssueID يولّد معرّفًا فريدًا لقضية جديدة.
func newIssueID() string {
	return "i-" + uuid.NewString()
}

// nowMillis يعيد الوقت الحالي بالمللي-ثانية.
func nowMillis() int64 {
	return time.Now().UnixMilli()
}
