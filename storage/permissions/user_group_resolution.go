package permstorage

const userGroupResolutionCTE = `
	WITH RECURSIVE
	member_group_ids(group_id) AS (
		SELECT gm.group_id
		FROM group_members gm
		WHERE gm.member_type = 'USER' AND gm.user_id = $1
		UNION
		SELECT gm.group_id
		FROM group_members gm
		JOIN member_group_ids mgi ON mgi.group_id = gm.member_group_id
		WHERE gm.member_type = 'GROUP'
	),
	user_group_ids(group_id) AS (
		SELECT group_id FROM member_group_ids
		UNION
		SELECT g.parent_group_id
		FROM groups g
		JOIN user_group_ids ugi ON ugi.group_id = g.id
		WHERE g.parent_group_id IS NOT NULL
	)
`
