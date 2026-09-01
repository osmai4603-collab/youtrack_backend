package sqlstore

import (
	"testing"
	"youtrack_backend/channels/model"
	"youtrack_backend/channels/model/fields"

	"github.com/stretchr/testify/assert"
)

func TestGetThreadColumns(t *testing.T) {
	s := &InboxStore{}

	t.Run("Empty tree returns all columns", func(t *testing.T) {
		cols, _ := s.getThreadColumns(nil)
		assert.Equal(t, 10, len(cols))
	})

	t.Run("Specific fields", func(t *testing.T) {
		tree := fields.Parse("id,read")
		cols, _ := s.getThreadColumns(tree)
		assert.Contains(t, cols, "id")
		assert.Contains(t, cols, "read")
		assert.NotContains(t, cols, "muted")
		assert.Equal(t, 2, len(cols))
	})

	t.Run("Subject fields", func(t *testing.T) {
		tree := fields.Parse("subject(text)")
		cols, _ := s.getThreadColumns(tree)
		assert.Contains(t, cols, "subject_text")
		assert.Contains(t, cols, "subject_target_id")
		// id should be included because of subject relation requirement in my implementation
		assert.Contains(t, cols, "id")
	})
}

func TestGetMessageColumns(t *testing.T) {
	s := &InboxStore{}

	t.Run("Specific message fields", func(t *testing.T) {
		tree := fields.Parse("text,timestamp")
		cols, _ := s.getMessageColumns(tree)
		assert.Contains(t, cols, "text")
		assert.Contains(t, cols, "timestamp")
		assert.NotContains(t, cols, "pseudo")
		assert.Equal(t, 2, len(cols))
	})
}

func TestMapColumnToThread(t *testing.T) {
	s := &InboxStore{}
	t.Run("Map ID", func(t *testing.T) {
		tr := &model.InboxThread{}
		s.mapColumnToThread(tr, "id", "thread-123")
		assert.Equal(t, "thread-123", tr.ID)
	})

	t.Run("Map Subject", func(t *testing.T) {
		tr := &model.InboxThread{}
		s.mapColumnToThread(tr, "subject_text", "Hello World")
		assert.NotNil(t, tr.Subject)
		assert.Equal(t, "Hello World", tr.Subject.Text)
	})
}
