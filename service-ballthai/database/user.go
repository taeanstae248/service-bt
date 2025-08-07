package database

import (
	"fmt"
	"time"
)

type User struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	FullName  *string    `json:"full_name,omitempty"`
	Role      string     `json:"role"`
	IsActive  bool       `json:"is_active"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    int       `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateUser สร้างผู้ใช้ใหม่
func CreateUser(username, email, passwordHash, fullName, role string) (*User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := DB.Exec(query, username, email, passwordHash, fullName, role)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetUserByID(int(id))
}

// GetUserByID ดึงข้อมูลผู้ใช้จาก ID
func GetUserByID(id int) (*User, error) {
	query := `
		SELECT id, username, email, full_name, role, is_active, created_at, updated_at, last_login
		FROM users 
		WHERE id = ?
	`

	var user User
	var lastLoginStr *string
	var createdAt, updatedAt string
	err := DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.Role,
		&user.IsActive, &createdAt, &updatedAt, &lastLoginStr,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	if createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		user.CreatedAt = &createdAtTime
	}
	if updatedAtTime, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
		user.UpdatedAt = &updatedAtTime
	}
	if lastLoginStr != nil {
		if lastLoginTime, err := time.Parse("2006-01-02 15:04:05", *lastLoginStr); err == nil {
			user.LastLogin = &lastLoginTime
		}
	}

	return &user, nil
}

// GetUserByUsername ดึงข้อมูลผู้ใช้จาก username
func GetUserByUsername(username string) (*User, error) {
	query := `
		SELECT id, username, email, full_name, role, is_active, created_at, updated_at, last_login
		FROM users 
		WHERE username = ? AND is_active = TRUE
	`

	var user User
	var lastLoginStr *string
	var createdAt, updatedAt string
	err := DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FullName, &user.Role,
		&user.IsActive, &createdAt, &updatedAt, &lastLoginStr,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps
	if createdAtTime, err := time.Parse("2006-01-02 15:04:05", createdAt); err == nil {
		user.CreatedAt = &createdAtTime
	}
	if updatedAtTime, err := time.Parse("2006-01-02 15:04:05", updatedAt); err == nil {
		user.UpdatedAt = &updatedAtTime
	}
	if lastLoginStr != nil {
		if lastLoginTime, err := time.Parse("2006-01-02 15:04:05", *lastLoginStr); err == nil {
			user.LastLogin = &lastLoginTime
		}
	}

	return &user, nil
}

// GetUserPasswordHash ดึง password hash สำหรับตรวจสอบรหัสผ่าน
func GetUserPasswordHash(username string) (string, error) {
	query := `SELECT password_hash FROM users WHERE username = ? AND is_active = TRUE`

	var passwordHash string
	err := DB.QueryRow(query, username).Scan(&passwordHash)
	return passwordHash, err
}

// UpdateLastLogin อัปเดต last_login timestamp
func UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := DB.Exec(query, userID)
	return err
}

// CreateSession สร้าง session ใหม่
func CreateSession(sessionID string, userID int, expiresAt time.Time) error {
	if DB == nil {
		return fmt.Errorf("database connection is nil")
	}

	query := `
		INSERT INTO sessions (id, user_id, expires_at)
		VALUES (?, ?, ?)
	`
	// Use local timezone for MySQL
	localTime := expiresAt.Local()
	_, err := DB.Exec(query, sessionID, userID, localTime)
	return err
}

// GetSession ดึงข้อมูล session
func GetSession(sessionID string) (*Session, error) {
	query := `
		SELECT id, user_id, expires_at, created_at
		FROM sessions 
		WHERE id = ?
	`

	var session Session
	var expiresAtStr, createdAtStr string
	err := DB.QueryRow(query, sessionID).Scan(
		&session.ID, &session.UserID, &expiresAtStr, &createdAtStr,
	)

	if err != nil {
		return nil, err
	}

	// Parse timestamps - support multiple formats
	timeFormats := []string{
		"2006-01-02T15:04:05Z07:00", // ISO 8601 with timezone
		"2006-01-02 15:04:05",       // MySQL datetime
		time.RFC3339,                // RFC3339 format
	}

	for _, format := range timeFormats {
		if expiresAtTime, err := time.Parse(format, expiresAtStr); err == nil {
			session.ExpiresAt = expiresAtTime
			break
		}
	}

	for _, format := range timeFormats {
		if createdAtTime, err := time.Parse(format, createdAtStr); err == nil {
			session.CreatedAt = createdAtTime
			break
		}
	}

	return &session, nil
}

// DeleteSession ลบ session
func DeleteSession(sessionID string) error {
	query := `DELETE FROM sessions WHERE id = ?`
	_, err := DB.Exec(query, sessionID)
	return err
}

// CleanExpiredSessions ลบ session ที่หมดอายุ
func CleanExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	_, err := DB.Exec(query)
	return err
}
