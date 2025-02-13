// Package dalerrs provides common error handling utilities for all domain-specific DALs,
// e.g. controller DAL and configuration DAL, which all connect to the same underlying DB
// and maintain the same interface guarantees
package dalerrs

import (
	stdsql "database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrConflict is returned by select methods in the DAL when a resource already exists.
	//
	// Its use will be documented in the corresponding methods.
	ErrConflict = errors.New("conflict")
	// ErrNotFound is returned by select methods in the DAL when no results are found.
	ErrNotFound = errors.New("not found")
	// ErrConstraint is returned by select methods in the DAL when a constraint is violated.
	ErrConstraint = errors.New("constraint violation")
)

func IsNotFound(err error) bool {
	return errors.Is(err, stdsql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows)
}

func TranslatePGError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%s: %w", strings.TrimSuffix(strings.TrimPrefix(pgErr.ConstraintName, pgErr.TableName+"_"), "_id_fkey"), ErrNotFound)
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%s: %w", pgErr.Message, ErrConflict)
		case pgerrcode.IntegrityConstraintViolation,
			pgerrcode.RestrictViolation,
			pgerrcode.NotNullViolation,
			pgerrcode.CheckViolation,
			pgerrcode.ExclusionViolation:
			return fmt.Errorf("%s: %w", pgErr.Message, ErrConstraint)
		default:
		}
	} else if IsNotFound(err) {
		return ErrNotFound
	}
	return err
}
