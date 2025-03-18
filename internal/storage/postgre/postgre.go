package postgre

import (
	"context"
	"errors"
	"fmt"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Storage struct {
	conn *pgx.Conn
}

func New(DBurl string) (*Storage, error) {
	const op = "storage.postgre.New"

	connect, err := pgx.Connect(context.Background(), DBurl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{conn: connect}, nil
}

func NewFromConn(conn *pgx.Conn) *Storage {
	return &Storage{conn: conn}
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgre.SaveURL"

	query := `
		INSERT INTO url(alias, url)
		VALUES($1, $2)
		RETURNING id
	`

	var id int64
	err := s.conn.QueryRow(context.Background(), query, alias, urlToSave).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return 0, storage.ErrURLExists
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.postgre.GetURL"

	query := `
		SELECT url FROM url 
		WHERE alias = $1
	`
	var result string

	err := s.conn.QueryRow(context.Background(), query, alias).Scan(&result)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *Storage) DeleteURL(alias string) error {
	//TODO
	return nil
}
