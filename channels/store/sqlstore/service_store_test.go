package sqlstore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"youtrack_backend/channels/model/fields"
)

func TestServiceColumns(t *testing.T) {
	t.Run("Empty tree returns all columns", func(t *testing.T) {
		ids, exprs := serviceColumns(nil)
		assert.Equal(t, 16, len(ids))
		assert.Equal(t, 16, len(exprs))
		assert.Contains(t, ids, "id")
		assert.Contains(t, ids, "trusted")
		assert.Contains(t, ids, "icon_url")
		assert.Contains(t, ids, "user_uri_pattern")
		assert.Contains(t, ids, "implicit_flow_enabled")
		// id يعود مباشرة، بينما الأعمدة النصية تُغلَّف بـ COALESCE
		assert.Equal(t, "id", exprs[0])
		assert.Equal(t, "trusted", exprs[7])
	})

	t.Run("Request24 fields", func(t *testing.T) {
		tree := fields.Parse("id,key,name,homeUrl,applicationName,vendor,version,trusted")
		ids, exprs := serviceColumns(tree)
		assert.Equal(t, 8, len(ids))
		assert.Equal(t, 8, len(exprs))
		assert.Contains(t, ids, "id")
		assert.Contains(t, ids, "key")
		assert.Contains(t, ids, "name")
		assert.Contains(t, ids, "home_url")
		assert.Contains(t, ids, "application_name")
		assert.Contains(t, ids, "vendor")
		assert.Contains(t, ids, "version")
		assert.Contains(t, ids, "trusted")
		assert.NotContains(t, ids, "icon_url")
		assert.NotContains(t, ids, "audience")
		assert.NotContains(t, ids, "immutable")
		// الأعمدة النصية تُغلَّف بـ COALESCE لتفادي أخطاء NULL
		assert.Contains(t, exprs, "COALESCE(home_url, '')")
		assert.Contains(t, exprs, "COALESCE(application_name, '')")
	})

	t.Run("Hub header request fields", func(t *testing.T) {
		tree := fields.Parse("applicationName,homeUrl,iconUrl,id,name,userUriPattern")
		ids, exprs := serviceColumns(tree)
		assert.Equal(t, 6, len(ids))
		assert.Equal(t, 6, len(exprs))
		assert.Contains(t, ids, "application_name")
		assert.Contains(t, ids, "home_url")
		assert.Contains(t, ids, "icon_url")
		assert.Contains(t, ids, "id")
		assert.Contains(t, ids, "name")
		assert.Contains(t, ids, "user_uri_pattern")
		assert.NotContains(t, ids, "key")
	})

	t.Run("Only id", func(t *testing.T) {
		tree := fields.Parse("id")
		ids, exprs := serviceColumns(tree)
		assert.Equal(t, 1, len(ids))
		assert.Equal(t, 1, len(exprs))
		assert.Equal(t, "id", ids[0])
	})

	t.Run("Unknown fields only fall back to all", func(t *testing.T) {
		tree := fields.Parse("bogusField")
		ids, exprs := serviceColumns(tree)
		assert.Equal(t, 16, len(ids))
		assert.Equal(t, 16, len(exprs))
	})
}
