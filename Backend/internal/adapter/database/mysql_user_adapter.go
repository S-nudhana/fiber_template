package adapter

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"os"
	"strconv"

	"github.com/google/uuid"
)

type MySQLUserAdapter struct {
	db *sql.DB
}

func NewMySQLUserAdapter(db *sql.DB) *MySQLUserAdapter {
	return &MySQLUserAdapter{db: db}
}

func (m *MySQLUserAdapter) CreateUser(email string, password string, firstName string, lastName string) (createStatus bool, err error) {
	var existingEmail string
	_ = m.db.QueryRow("SELECT user_email FROM users WHERE user_email = ?", email).Scan(&existingEmail)
	if existingEmail != "" {
		return false, errors.New("email already exists")
	}
	userId := uuid.New()
	cost, err := strconv.Atoi(os.Getenv("BCRYPT_COST"))
	if err != nil {
		cost = bcrypt.DefaultCost
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return false, errors.New("failed to hash password")
	}

	tx, err := m.db.Begin()
	if err != nil {
		return false, err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		INSERT INTO users
		(user_id, user_email, user_firstName, user_lastName, user_authType)
		VALUES (?, ?, ?, ?, ?)
	`, userId, email, firstName, lastName, "local")
	if err != nil {
		return false, err
	}
	_, err = tx.Exec(`
		INSERT INTO users_local (user_id, user_password)
		VALUES (?, ?)
	`, userId, hashPassword)
	if err != nil {
		return false, err
	}
	err = tx.Commit()
	if err != nil {
		return false, errors.New("failed to create user")
	}
	return true, nil
}

func (m *MySQLUserAdapter) OAuthAuthenticateUser(
	email string,
	provider string,
	firstName string,
	lastName string,
) (bool, string, error) {
	var userUID string
	err := m.db.QueryRow(`
		SELECT u.user_id
		FROM users u
		JOIN users_oauth uo ON u.user_id = uo.user_id
		WHERE u.user_email = ? AND uo.user_provider = ?
	`, email, provider).Scan(&userUID)
	if err == nil {
		return true, userUID, nil
	}

	if err != sql.ErrNoRows {
		return false, "", err
	}
	newUID := uuid.New().String()
	oAuthUID := newUID + ":" + provider

	tx, err := m.db.Begin()
	if err != nil {
		return false, "", err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		INSERT INTO users
		(user_id, user_email, user_firstName, user_lastName, user_authType)
		VALUES (?, ?, ?, ?, ?)
	`, oAuthUID, email, firstName, lastName, "oauth")
	if err != nil {
		return false, "", err
	}
	_, err = tx.Exec(`
		INSERT INTO users_oauth (user_id, user_provider)
		VALUES (?, ?)
	`, oAuthUID, provider)
	if err != nil {
		return false, "", err
	}
	err = tx.Commit()
	if err != nil {
		return false, "", err
	}
	return true, oAuthUID, nil
}

func (m *MySQLUserAdapter) AuthenticateUser(email string, password string) (authStatus bool, uid string, err error) {
	var storedPassword string
	var userId uuid.UUID

	err = m.db.QueryRow(`
		SELECT u.user_id, ul.user_password 
		FROM users AS u 
		JOIN users_local AS ul ON u.user_id = ul.user_id  
		WHERE u.user_email = ?
	`, email).Scan(&userId, &storedPassword)
	if err != nil {
		return false, "", errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return false, "", errors.New("password does not match")
	}

	return true, userId.String(), nil
}

func (m *MySQLUserAdapter) RemoveUser(uid string) (deleteStatus bool, err error) {
	userId, err := uuid.Parse(uid)
	if err != nil {
		return false, errors.New("invalid user ID")
	}
	_, err = m.db.Exec("DELETE FROM users WHERE user_id = ?", userId)
	if err != nil {
		return false, errors.New("failed to delete user")
	}
	return true, nil
}

func (m *MySQLUserAdapter) UpdateUserInfo(uid string, firstName string, lastName string) (updateStatus bool, err error) {
	userId, err := uuid.Parse(uid)
	if err != nil {
		return false, errors.New("invalid user ID")
	}

	result, err := m.db.Exec("UPDATE users SET user_firstName = ?, user_lastName = ? WHERE user_id = ?", firstName, lastName, userId)
	if err != nil {
		return false, errors.New("failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 || err != nil {
		return false, errors.New("user not found")
	}
	return true, nil
}
