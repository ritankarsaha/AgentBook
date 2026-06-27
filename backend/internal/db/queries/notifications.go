package queries

const (
	GetNotifications = `
		SELECT
		  n.id, n.type, n.actor_user_id, n.actor_agent_id, n.post_id, n.read, n.created_at,
		  COALESCE(au.handle, aa.handle)             AS actor_handle,
		  COALESCE(au.display_name, aa.display_name) AS actor_display_name,
		  COALESCE(au.avatar_url, aa.avatar_url)     AS actor_avatar_url
		FROM notifications n
		LEFT JOIN users  au ON n.actor_user_id  = au.id
		LEFT JOIN agents aa ON n.actor_agent_id = aa.id
		WHERE (n.recipient_user_id  IS NOT DISTINCT FROM $1
		    OR n.recipient_agent_id IS NOT DISTINCT FROM $2)
		  AND ($3::timestamptz IS NULL OR n.created_at < $3)
		ORDER BY n.created_at DESC
		LIMIT $4`

	MarkNotificationsRead = `
		UPDATE notifications
		SET read = true
		WHERE (recipient_user_id  IS NOT DISTINCT FROM $1
		    OR recipient_agent_id IS NOT DISTINCT FROM $2)
		  AND read = false`
)
