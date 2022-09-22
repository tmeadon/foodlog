package data

import (
	"database/sql"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id           int    `db:"id"`
	Username     string `db:"username"`
	PasswordHash string `db:"password_hash"`
	IsAdmin      bool   `db:"is_admin"`
}

func NewUser(username string, password string) (*User, error) {
	user := User{
		Username: username,
		IsAdmin:  false,
	}
	user.SetPassword(password)
	return &user, nil
}

func (u *User) ValidatePassword(password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil, err
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.PasswordHash = string(hash)
	return nil
}

func GetUsers() ([]User, error) {
	var users []User
	err := DB.Select(&users, "select * from user order by username asc")
	return users, err
}

func GetUserById(id int) (*User, error) {
	var user User
	row := DB.QueryRowx("select id, username, is_admin from user where id = ?", id)
	err := row.StructScan(&user)
	return &user, err
}

func GetUserByIdWithPasswordHash(id int) (*User, error) {
	var user User
	row := DB.QueryRowx("select * from user where id = ?", id)
	err := row.StructScan(&user)
	return &user, err
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	row := DB.QueryRowx("select id, username, is_admin from user where username = ?", username)
	err := row.StructScan(&user)
	return &user, err
}

func GetUserByUsernameWithPasswordHash(username string) (*User, error) {
	var user User
	row := DB.QueryRowx("select id, username, is_admin, password_hash from user where username = ?", username)
	err := row.StructScan(&user)
	return &user, err
}

func SaveUser(user *User) error {
	_, err := GetUserById(user.Id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := insertUser(user)
			return err
		} else {
			return err
		}
	}

	return updateUser(user)
}

func insertUser(user *User) error {
	_, err := DB.NamedExec(`insert into user (username, password_hash, is_admin) values (:username, :passwordhash, :isadmin)`,
		map[string]interface{}{
			"username":     user.Username,
			"passwordhash": user.PasswordHash,
			"isadmin":      user.IsAdmin,
		})
	return err
}

func updateUser(user *User) error {
	_, err := DB.NamedExec(`update user set username=:username, password_hash=:passwordhash, is_admin=:isadmin where id = :id`,
		map[string]interface{}{
			"username":     user.Username,
			"passwordhash": user.PasswordHash,
			"isadmin":      user.IsAdmin,
			"id":           user.Id,
		})
	return err
}

func DeleteUser(user *User) error {
	_, err := DB.Exec("delete from user where id = ?", user.Id)
	return err
}
