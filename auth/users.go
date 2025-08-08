package auth

import (
	"database/sql"
	"errors"
	"time"

	"lod2/db"

	"github.com/mattn/go-sqlite3"
	"go.jetify.com/typeid"
)

// authUsers has rows:
// userId TEXT
// userName TEXT
// userPasswordHash TEXT
// inviteId TEXT -- the unique invite code this user used to register
// createdAt INTEGER -- when the user was created, unix time

// Creates a user with the provided username and password.
func createUser(tx *sql.Tx, username string, password string, roles []Role) (string, error) {
	userId, _ := typeid.WithPrefix("user")
	passwordHash, err := hashPassword(password)

	if err != nil {
		return "", err
	}

	_, err = tx.Exec("INSERT INTO authUsers (userId, userName, userPasswordHash, createdAt) VALUES (?, ?, ?, ?)", userId, username, passwordHash, time.Now().Unix())

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return "", errors.New("this username is already taken")
			}
		}

		return "", err
	}

	inviteId, _ := typeid.WithPrefix("inv")

	_, err = tx.Exec("INSERT INTO authInvites (inviteId, userId, issuedAt, inviteLimit) VALUES (?, ?, ?, ?)", inviteId, userId, time.Now().Unix(), 5)

	if err != nil {
		return "", err
	}

	err = addRoles(tx, userId.String(), roles)

	if err != nil {
		return "", err
	}

	return userId.String(), nil
}

// Returns the user ID, or an error if the user does not exist or the password is incorrect.
func getUserLogin(username string, password string) (string, error) {
	var userId string
	var passwordHash string

	err := db.DB.QueryRow("SELECT userId, userPasswordHash FROM authUsers WHERE userName = ?", username).Scan(&userId, &passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid username")
		}

		return "", err
	}

	passwordValid := verifyPassword(passwordHash, password)

	if !passwordValid {
		return "", errors.New("invalid password")
	}

	return string(userId), nil
}

// Returns an error if the userId does not exist or their password is incorrect.
func verifyUserPassword(userId string, password string) error {
	var passwordHash string

	err := db.DB.QueryRow("SELECT userPasswordHash FROM authUsers WHERE userId = ?", userId).Scan(&passwordHash)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid user id")
		}

		return err
	}

	passwordValid := verifyPassword(passwordHash, password)

	if !passwordValid {
		return errors.New("invalid password")
	}

	return nil
}

// User-facing change password function. Used by a user to change their own password.
func ChangePassword(userId string, currentPassword string, newPassword string, newPasswordVerify string) error {
	err := verifyUserPassword(userId, currentPassword)

	if err != nil {
		return errors.New("invalid current password")
	}

	if newPassword != newPasswordVerify {
		return errors.New("new passwords do not match")
	}

	newPasswordHash, err := hashPassword(newPassword)

	if err != nil {
		return err
	}

	_, err = db.DB.Exec("UPDATE authUsers SET userPasswordHash = ? WHERE userId = ?", newPasswordHash, userId)

	if err != nil {
		return err
	}

	return nil
}

type UserSessionInfo struct {
	UserId           string
	Username         string
	LastLogin        time.Time
	LastActivity     time.Time
	CreatedAt        time.Time
	SessionCount     int
	InvitesRemaining int
	InvitedByUserId  *string
	Roles            []Role
}

func AdminGetAllUsers() ([]UserSessionInfo, error) {
	rows, err := db.DB.Query(`
	SELECT
    authUsers.userId,
    authUsers.userName,
    COALESCE(MAX(authSessions.issuedAt), 0) AS lastLogin,
    COALESCE(MAX(authSessions.refreshedAt), 0) AS lastActivity,
    authUsers.createdAt,
    COALESCE(COUNT(CASE WHEN authSessions.expiresAt > ? THEN 1 ELSE NULL END), 0) AS sessionCount,
    COALESCE(invites.inviteLimitTotal, 0) - COUNT(authUsers.inviteId) AS invitesRemaining
	FROM authUsers
	LEFT JOIN authSessions ON authUsers.userId = authSessions.userId
	LEFT JOIN (
    -- Aggregate invite limits per user first
    SELECT userId, SUM(inviteLimit) AS inviteLimitTotal
    FROM authInvites
    GROUP BY userId
  ) AS invites ON invites.userId = authUsers.userId
  GROUP BY authUsers.userId, authUsers.userName, invites.inviteLimitTotal;`, time.Now().Unix())

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users []UserSessionInfo

	for rows.Next() {
		var userId string
		var userName string
		var lastLogin int64
		var lastActivity int64
		var createdAt int64
		var sessionCount int
		var invitesRemaining int

		err := rows.Scan(&userId, &userName, &lastLogin, &lastActivity, &createdAt, &sessionCount, &invitesRemaining)

		if err != nil {
			return nil, err
		}

		if lastActivity == 0 {
			lastActivity = lastLogin
		}

		users = append(users, UserSessionInfo{
			UserId:           userId,
			Username:         userName,
			LastLogin:        time.Unix(lastLogin, 0),
			LastActivity:     time.Unix(lastActivity, 0),
			CreatedAt:        time.Unix(createdAt, 0),
			SessionCount:     sessionCount,
			InvitesRemaining: invitesRemaining,
		})
	}

	return users, nil
}

func AdminGetUserIdByUsername(tx *sql.Tx, username string) (string, error) {
	row := tx.QueryRow(`
		SELECT
			authUsers.userId
		FROM authUsers
		WHERE authUsers.userName = ?`, username)

	if row == nil {
		return "", errors.New("invalid username")
	}

	var userId string
	err := row.Scan(&userId)

	if err != nil {
		return "", err
	}

	return userId, nil
}

func AdminGetUserById(userId string) (UserSessionInfo, error) {
	row := db.DB.QueryRow(`
        WITH invite_totals AS (
            SELECT userId, SUM(inviteLimit) AS inviteLimitTotal
            FROM authInvites
            GROUP BY userId
        )
        SELECT
            u.userId,
            u.userName,
            COALESCE(MAX(s.issuedAt), 0) AS lastLogin,
            u.createdAt,
            COALESCE(COUNT(CASE WHEN s.expiresAt > ? THEN 1 END), 0) AS sessionCount,
            COALESCE(it.inviteLimitTotal, 0) - COUNT(u.inviteId) AS invitesRemaining,
            i.userId AS invitedByUserId
        FROM authUsers AS u
        LEFT JOIN authSessions AS s ON u.userId = s.userId
        LEFT JOIN invite_totals AS it ON it.userId = u.userId
        LEFT JOIN authInvites AS i ON i.inviteId = u.inviteId
        WHERE u.userId = ?
        GROUP BY u.userId, u.userName, u.createdAt, it.inviteLimitTotal, i.userId
    `, time.Now().Unix(), userId)

	var userName string
	var lastLogin int64
	var createdAt int64
	var sessionCount int
	var invitesRemaining int
	var invitedByUserId *string

	err := row.Scan(&userId, &userName, &lastLogin, &createdAt, &sessionCount, &invitesRemaining, &invitedByUserId)

	if err != nil {
		if err == sql.ErrNoRows {
			return UserSessionInfo{}, errors.New("invalid user id")
		}
		return UserSessionInfo{}, err
	}

	// Fetch and map roles for this user
	roles, err := GetUserRoles(userId)
	if err != nil {
		return UserSessionInfo{}, err
	}

	return UserSessionInfo{
		UserId:           userId,
		Username:         userName,
		LastLogin:        time.Unix(lastLogin, 0),
		CreatedAt:        time.Unix(createdAt, 0),
		SessionCount:     sessionCount,
		InvitesRemaining: invitesRemaining,
		InvitedByUserId:  invitedByUserId,
		Roles:            roles,
	}, nil
}
