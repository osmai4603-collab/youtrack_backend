CREATE TABLE security_filter_fields (
    id VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    field_type VARCHAR(100) DEFAULT 'SecurityFilterField'
);

-- إدراج البيانات للطلب رقم 30 (ProjectPeopleResponse)
INSERT INTO security_filter_fields (id, name, entity_type, field_type) VALUES
('role.scope', 'Scope', 'ProjectPeopleResponse', 'SecurityFilterField'),
('role.role', 'Role', 'ProjectPeopleResponse', 'SecurityFilterField'),
('role.permission', 'Permission', 'ProjectPeopleResponse', 'SecurityFilterField');

-- يمكن إضافة كيانات أخرى كما ظهرت في hh.json
INSERT INTO security_filter_fields (id, name, entity_type, field_type) VALUES
('user.login', 'Login', 'User', 'SecurityFilterField'),
('user.name', 'Name', 'User', 'SecurityFilterField'),
('group.name', 'Name', 'UserGroup', 'SecurityFilterField'),
('role.name', 'Name', 'Role', 'SecurityFilterField');
