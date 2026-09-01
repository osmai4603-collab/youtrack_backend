package request

import (
	"context"
	"youtrack_backend/channels/model/shared/mlog"
)

// CTX يمثل الواجهة المعيارية لسياق الطلبات عبر كافة طبقات المنصة، مستوحاة من Mattermost request.CTX.
type CTX interface {
	Context() context.Context
	RequestId() string
	IPAddress() string
	Path() string
	UserAgent() string
	UserID() string
	SessionToken() string

	WithContext(ctx context.Context) CTX
	WithRequestId(string) CTX
	WithIPAddress(string) CTX
	WithPath(string) CTX
	WithUserAgent(string) CTX
	WithUserID(string) CTX
	WithSessionToken(string) CTX
}

// Context هو التطبيق الملموس لواجهة CTX.
type Context struct {
	context      context.Context
	requestId    string
	ipAddress    string
	path         string
	userAgent    string
	userId       string
	sessionToken string
	logger       mlog.LoggerIFace
}

func NewContext(ctx context.Context, requestId, ipAddress, path, userAgent string) *Context {
	return &Context{
		requestId: requestId,
		ipAddress: ipAddress,
		path:      path,
		userAgent: userAgent,
		context:   ctx,
	}
}

// NewContext ينشئ سياق طلب جديد مع الحقول الأساسية.
// func NewContext(ctx context.Context, requestId, ipAddress, path, userAgent string) *Context {
// 	if ctx == nil {
// 		ctx = context.Background()
// 	}
// 	return &Context{
// 		context:   ctx,
// 		requestId: requestId,
// 		ipAddress: ipAddress,
// 		path:      path,
// 		userAgent: userAgent,
// 	}
// }

// EmptyContext ينشئ سياق طلب فارغاً للاستخدامات العامة والمهام الخلفية.
func EmptyContext(logger mlog.LoggerIFace) *Context {
	return &Context{
		logger:  logger,
		context: context.Background(),
	}
}

// TestContext ينشئ سياقاً مخصصاً للاختبارات الأحادية.
func TestContext() *Context {
	return &Context{
		context:   context.Background(),
		requestId: "test-request-id",
		userId:    "test-user-id",
	}
}

func (c *Context) clone() *Context {
	cCopy := *c
	return &cCopy
}

func (c *Context) Context() context.Context {
	if c.context == nil {
		return context.Background()
	}
	return c.context
}

func (c *Context) RequestId() string    { return c.requestId }
func (c *Context) IPAddress() string    { return c.ipAddress }
func (c *Context) Path() string         { return c.path }
func (c *Context) UserAgent() string    { return c.userAgent }
func (c *Context) UserID() string       { return c.userId }
func (c *Context) SessionToken() string { return c.sessionToken }

func (c *Context) WithContext(ctx context.Context) CTX {
	n := c.clone()
	n.context = ctx
	return n
}

func (c *Context) WithRequestId(id string) CTX {
	n := c.clone()
	n.requestId = id
	return n
}

func (c *Context) WithIPAddress(ip string) CTX {
	n := c.clone()
	n.ipAddress = ip
	return n
}

func (c *Context) WithPath(p string) CTX {
	n := c.clone()
	n.path = p
	return n
}

func (c *Context) WithUserAgent(ua string) CTX {
	n := c.clone()
	n.userAgent = ua
	return n
}

func (c *Context) WithUserID(uid string) CTX {
	n := c.clone()
	n.userId = uid
	return n
}

func (c *Context) WithSessionToken(tok string) CTX {
	n := c.clone()
	n.sessionToken = tok
	return n
}
