package model

// ProjectTeamAndLeader هو مخطط استجابة مخصص للطلب #27
// (GET /api/admin/projects/{id}?fields=team(name,users(...)),leader(id)).
// هذا المخطط مستقل تماماً عن model.Project ولا يلمس طلبات المشاريع الأخرى.
type ProjectTeamAndLeader struct {
	Leader *User27 `json:"leader"`
	Team   *Team27 `json:"team"`
	Type   string  `json:"$type"`
}

// Normalize يملأ $type الافتراضي لكائن الجذر.
func (p *ProjectTeamAndLeader) Normalize() {
	p.Type = "Project"
}

// User27 هو مستخدم مبسّط مخصص للطلب #27 بحقول قابلة للـ null.
// email و avatarUrl من نوع pointer حتى يمكن إخراجهما null في JSON
// (كما في استجابة request27.txt لحساب الضيف).
type User27 struct {
	ID        string  `json:"id"`
	Login     string  `json:"login"`
	Name      string  `json:"name,omitempty"`
	AvatarURL *string `json:"avatarUrl"`
	Email     *string `json:"email"`
	Type      string  `json:"$type"`
}

// Normalize يملأ $type الافتراضي للمستخدم.
func (u *User27) Normalize() {
	u.Type = "User"
}

// Team27 هو فريق مبسّط مخصص للطلب #27.
type Team27 struct {
	Name  string    `json:"name"`
	Users []*User27 `json:"users,omitempty"`
	Type  string    `json:"$type"`
}

// Normalize يملأ $type الافتراضي للفريق.
func (t *Team27) Normalize() {
	t.Type = "ProjectTeam"
}
