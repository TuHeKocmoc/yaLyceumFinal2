package repository

import (
	"database/sql"
	"errors"

	"github.com/mattn/go-sqlite3"

	"github.com/TuHeKocmoc/yalyceumfinal2/internal/db"
	"github.com/TuHeKocmoc/yalyceumfinal2/internal/model"
)

func CreateUser(login, passwordHash string) error {
	query := `INSERT INTO users (login, password_hash) VALUES (?, ?)`
	_, err := db.GlobalDB.Exec(query, login, passwordHash)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return errors.New("user already exists")
			}
		}
		return err
	}
	return nil
}

func GetUserByLogin(login string) (*model.User, error) {
	query := `SELECT id, login, password_hash FROM users WHERE login = ?`
	row := db.GlobalDB.QueryRow(query, login)

	var u model.User
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func GetUserByID(id int64) (*model.User, error) {
	query := `SELECT id, login, password_hash FROM users WHERE id = ?`
	row := db.GlobalDB.QueryRow(query, id)

	var u model.User
	err := row.Scan(&u.ID, &u.Login, &u.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
