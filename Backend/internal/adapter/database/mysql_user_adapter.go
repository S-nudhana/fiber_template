package adapter

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"os"
	"strconv"
)

type MySQLUserAdapter struct {
	db *sql.DB
}

func NewMySQLUserAdapter(db *sql.DB) *MySQLUserAdapter {
	return &MySQLUserAdapter{db: db}
}

func (m *MySQLUserAdapter) CreateUser(email string, password string, firstname string, lastname string) (status bool, err error) {
	var existingEmail string
	_ = m.db.QueryRow("SELECT email FROM users WHERE email = ?", email).Scan(&existingEmail)
	if existingEmail != "" {
		return false, errors.New("email already exists")
	}

	cost, _ := strconv.Atoi(os.Getenv("BCRYPT_COST"))
	hashPassword, _ := bcrypt.GenerateFromPassword([]byte(password), cost)
	_, err = m.db.Exec("INSERT INTO users (email, password, firstname, lastname) VALUES (?, ?, ?, ?)", email, hashPassword, firstname, lastname)
	if err != nil {
		return false, errors.New("failed to create user")
	}  
	return true, nil
}

func (m *MySQLUserAdapter) AuthenticateUser(email string, password string) (status bool, uid string, err error) {
	var storedPassword string
	var userID string
	err = m.db.QueryRow("SELECT uid, password FROM users WHERE email = ?", email).Scan(&userID, &storedPassword)
	if err != nil {
		return false, "", errors.New("user not found")
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return false, "", errors.New("password does not match")
	}
	return true, userID, nil
}
