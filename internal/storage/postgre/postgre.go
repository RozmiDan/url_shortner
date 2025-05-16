package postgre

import (
	"context"
	"errors"
	"fmt"

	"github.com/RozmiDan/url_shortener/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(DBurl string) (*Storage, error) {
	const op = "storage.postgre.New"

	// Создаем конфигурацию пула
	config, err := pgxpool.ParseConfig(DBurl)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	config.MaxConns = 50
	config.MinConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgre.SaveURL"

	query := `
		INSERT INTO url(alias, url)
		VALUES($1, $2)
		RETURNING id
	`

	var id int64
	err := s.pool.QueryRow(context.Background(), query, alias, urlToSave).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// TODO
			if pgErr.Code == "23505" {
				return 0, storage.ErrAliasExists
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

	err := s.pool.QueryRow(context.Background(), query, alias).Scan(&result)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (s *Storage) DeleteURL(alias string) error {
	const op = "storage.postgre.DeleteURL"

	query := `
		DELETE FROM url
		WHERE alias = $1;
	`

	cmdTag, err := s.pool.Exec(context.Background(), query, alias)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return storage.ErrAliasNotFound
	}

	return nil
}

func (s *Storage) UpdateURL(currAlias string, newAlias string) error {
	const op = "storage.postgre.UpdateURL"

	query := `
		UPDATE url 
		SET alias = $1
		WHERE alias = $2;
	`

	cmdTag, err := s.pool.Exec(context.Background(), query, newAlias, currAlias)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return storage.ErrAliasExists
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return storage.ErrAliasNotFound
	}

	return nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
