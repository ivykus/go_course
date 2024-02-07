package mdb

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

func TryCreate(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE emails (
			id  INTEGER PRIMARY KEY,
			email TEXT UNIQUE,
			confirmed_at INTEGER,
			opt_out INTEGER
		);
	`)
	if err != nil {
		var sqlErr sqlite3.Error
		if errors.As(err, &sqlErr) {
			if sqlErr.Code != sqlite3.ErrError {
				log.Fatal(sqlErr)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func EmailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var emailEntry EmailEntry
	var confirmedAt int64
	err := row.Scan(&emailEntry.Id, &emailEntry.Email, &confirmedAt, &emailEntry.OptOut)
	if err != nil {
		log.Printf("EmailEntryFromRow: %v", err)
		return nil, err
	}
	t := time.Unix(confirmedAt, 0)
	emailEntry.ConfirmedAt = &t
	return &emailEntry, nil
}

func CreateEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		INSERT INTO emails(email, confirmed_at, opt_out) 
		VALUES (?, 0, false)`, email)
	if err != nil {
		log.Printf("CreateEmail: %v", err)
		return err
	}
	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE email = ?`, email)

	if err != nil {
		log.Printf("GetEmail: %v", err)
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		return EmailEntryFromRow(rows)
	}

	return nil, nil
}

func UpdateEmail(db *sql.DB, emailEntry *EmailEntry) error {
	t := emailEntry.ConfirmedAt.Unix()
	_, err := db.Exec(`
		INSERT INTO emails(email, confirmed_at, opt_out)
		VALUES(?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
		confirmed_at = ?, opt_out = ?`,
		emailEntry.Email, t, emailEntry.OptOut,
		t, emailEntry.OptOut)

	if err != nil {
		log.Printf("UpdateEmail: %v", err)
		return err
	}
	return nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		UPDATE emails
		SET opt_out = true
		WHERE email = ?`, email)
	if err != nil {
		log.Printf("DeleteEmail: %v", err)
		return err
	}
	return nil
}

type GetEmailBatchQueryParams struct {
	Page  int
	Count int
}

func GetEmailBatch(db *sql.DB, params GetEmailBatchQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE opt_out = false
		ORDER BY id ASC
		LIMIT ? OFFSET ?`, params.Count, (params.Page-1)*params.Count)

	if err != nil {
		log.Printf("GetEmailBatch: %v", err)
		return empty, err
	}
	defer rows.Close()

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := EmailEntryFromRow(rows)
		if err != nil {
			log.Printf("GetEmailBatch: %v", err)
			return nil, err
		}
		emails = append(emails, *email)
	}

	return emails, nil
}
