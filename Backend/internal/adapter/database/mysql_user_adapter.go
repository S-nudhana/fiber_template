package adapter

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"os"
	"strconv"
	"time"
	
	"github.com/google/uuid"
)

type MySQLUserAdapter struct {
	db *sql.DB
}

func NewMySQLUserAdapter(db *sql.DB) *MySQLUserAdapter {
	return &MySQLUserAdapter{db: db}
}

func (m *MySQLUserAdapter) CreateUser(email string, password string, firstname string, lastname string) (createStatus bool, err error) {
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

	now := time.Now()
	_, err = m.db.Exec("INSERT INTO users (user_id, user_email, user_password, user_firstname, user_lastname, user_createDate, user_modifyDate) VALUES (?, ?, ?, ?, ?, ?, ?)", userId, email, hashPassword, firstname, lastname, now, now)
	if err != nil {
		return false, errors.New("failed to create user")
	}
	return true, nil
}

func (m *MySQLUserAdapter) OAuthAuthenticateUser(email string, provider string, firstName string, lastName string) (authStatus bool, uid string, err error){
	var userId string

	err = m.db.QueryRow("SELECT uid, user_provider FROM users_oAuth WHERE user_email = ?", email).Scan(&userId)
	if err == nil {
		return true, userId, nil
	}

	new_userId := uuid.New()
	oAuthUID := new_userId.String() + ":" + provider
	now := time.Now()

	_, err = m.db.Exec("INSERT INTO users_oAuth (uid, user_email, user_provider, user_firstname, user_lastname, user_modifyDate) VALUES (?, ?, ?, ?, ?, ?)", oAuthUID, email, provider, firstName, lastName, now)
	if err != nil {
		return false, "", errors.New("failed to create user")
	}
	return true, oAuthUID, nil
}

func (m *MySQLUserAdapter) AuthenticateUser(email string, password string) (authStatus bool, uid string, err error) {
	var storedPassword string
	var userId uuid.UUID

	err = m.db.QueryRow("SELECT user_id, user_password FROM users WHERE user_email = ?", email).Scan(&userId, &storedPassword)
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

func (m *MySQLUserAdapter) UpdateUserInfo(uid string, firstname string, lastname string) (updateStatus bool, err error) {
	userId, err := uuid.Parse(uid)
	if err != nil {
		return false, errors.New("invalid user ID")
	}

	now := time.Now()
	result, err := m.db.Exec("UPDATE users SET user_firstname = ?, user_lastname = ?, user_modifyDate = ? WHERE user_id = ?", firstname, lastname, now, userId)
	if err != nil {
		return false, errors.New("failed to update user")
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 || err != nil {
		return false, errors.New("user not found")
	}
	return true, nil
}