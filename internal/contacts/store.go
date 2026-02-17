package contacts

import (
	"database/sql"
	"time"

	_ "modernc.org/sqlite"
)

type Contact struct {
	Phone     string `json:"phone"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	FirstSeen string `json:"firstSeen"`
	LastSeen  string `json:"lastSeen"`
}

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS contacts (
			phone      TEXT PRIMARY KEY,
			name       TEXT DEFAULT '',
			source     TEXT DEFAULT 'message',
			first_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen  DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Upsert(phone, name, source string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`
		INSERT INTO contacts (phone, name, source, first_seen, last_seen)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(phone) DO UPDATE SET
			name = CASE WHEN excluded.name != '' THEN excluded.name ELSE contacts.name END,
			last_seen = excluded.last_seen
	`, phone, name, source, now, now)
	return err
}

func (s *Store) GetAll(limit, offset int) ([]Contact, int, error) {
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(
		"SELECT phone, name, source, first_seen, last_seen FROM contacts ORDER BY last_seen DESC LIMIT ? OFFSET ?",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.Phone, &c.Name, &c.Source, &c.FirstSeen, &c.LastSeen); err != nil {
			return nil, 0, err
		}
		contacts = append(contacts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return contacts, total, nil
}

func (s *Store) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM contacts").Scan(&count)
	return count, err
}

func (s *Store) Close() error {
	return s.db.Close()
}
