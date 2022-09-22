package data

import (
	"database/sql"
	"errors"
	"time"
)

type LogEntry struct {
	Id     int       `db:"id"`
	UserId int       `db:"user_id"`
	Time   time.Time `db:"time"`
	Food   string    `db:"food"`
	Notes  string    `db:"notes"`
}

func GetEntriesByUser(userid int) ([]LogEntry, error) {
	var entries []LogEntry
	err := DB.Select(&entries, "select * from logentry where user_id = ? order by time desc", userid)
	return entries, err
}

func GetEntryById(id int) (*LogEntry, error) {
	var entry LogEntry
	row := DB.QueryRowx("select * from logentry where id = ?", id)
	err := row.StructScan(&entry)
	return &entry, err
}

func SaveEntry(entry *LogEntry) error {
	_, err := GetEntryById(entry.Id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err := insertEntry(entry)
			return err
		} else {
			return err
		}
	}

	return updateEntry(entry)
}

func insertEntry(entry *LogEntry) error {
	_, err := DB.NamedExec(`insert into logentry (user_id, time, food, notes) values (:user_id, :time, :food, :notes)`,
		map[string]interface{}{
			"user_id": entry.UserId,
			"time":    entry.Time,
			"food":    entry.Food,
			"notes":   entry.Notes,
		})
	return err
}

func updateEntry(entry *LogEntry) error {
	_, err := DB.NamedExec(`update user set user_id=:user_id, time=:time, food=:food, notes=:notes where id = :id`,
		map[string]interface{}{
			"user_id": entry.UserId,
			"time":    entry.Time,
			"food":    entry.Food,
			"notes":   entry.Notes,
			"id":      entry.Id,
		})
	return err
}

func DeleteEntry(entry *LogEntry) error {
	_, err := DB.Exec("delete from logentry where id = ?", entry.Id)
	return err
}
