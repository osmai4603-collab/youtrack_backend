package sqlstore

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"youtrack_backend/internal/store"
)

// SqlStore هو التنفيذ الملموس لواجهة store.Store باستخدام pgx.
type SqlStore struct {
	db           *pgxpool.Pool
	users        *UserStore
	projects     *ProjectStore
	issues       *IssueStore
	admin        *AdminStore
	inbox        *InboxStore
	savedQueries *SavedQueryStore
	search       *SearchStore
}

// New ينشئ SqlStore ويتحقق من الاتصال بقاعدة البيانات.
func New(dsn string) (*SqlStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}
	log.Println("Connected to PostgreSQL database successfully")

	s := &SqlStore{db: db}
	s.users = &UserStore{db: db}
	s.projects = &ProjectStore{db: db}
	s.issues = &IssueStore{db: db}
	s.admin = &AdminStore{db: db}
	s.inbox = &InboxStore{db: db}
	s.savedQueries = &SavedQueryStore{db: db}
	s.search = &SearchStore{db: db}
	return s, nil
}

// Close يغلق مجمع الاتصالات.
func (s *SqlStore) Close() {
	s.db.Close()
}

func (s *SqlStore) Users() store.UserStore              { return s.users }
func (s *SqlStore) Projects() store.ProjectStore        { return s.projects }
func (s *SqlStore) Issues() store.IssueStore            { return s.issues }
func (s *SqlStore) Admin() store.AdminStore             { return s.admin }
func (s *SqlStore) Inbox() store.InboxStore             { return s.inbox }
func (s *SqlStore) SavedQueries() store.SavedQueryStore { return s.savedQueries }
func (s *SqlStore) Search() store.SearchStore           { return s.search }

// userColumns يمثّل أسماء أعمدة جدول users (للاستخدام في SELECT).
const userColumns = `id, login, email, full_name, name, avatar_url, user_type_id,
	is_email_verified, guest, online, banned, can_read_profile, is_locked, ring_id`
