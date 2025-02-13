// Package dal provides a data abstraction layer for managing module configurations
package dal

import (
	"context"

	"github.com/alecthomas/types/optional"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/TBD54566975/ftl/db/dalerrs"
)

type DAL struct {
	db sql.DBI
}

func New(ctx context.Context, pool *pgxpool.Pool) (*DAL, error) {
	dal := &DAL{db: sql.NewDB(pool)}
	return dal, nil
}

func (d *DAL) GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error) {
	b, err := d.db.GetModuleConfiguration(ctx, module, name)
	if err != nil {
		return nil, dalerrs.TranslatePGError(err)
	}
	return b, nil
}

func (d *DAL) SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error {
	err := d.db.SetModuleConfiguration(ctx, module, name, value)
	return dalerrs.TranslatePGError(err)
}

func (d *DAL) UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error {
	err := d.db.UnsetModuleConfiguration(ctx, module, name)
	return dalerrs.TranslatePGError(err)
}

func (d *DAL) ListModuleConfiguration(ctx context.Context) ([]sql.ModuleConfiguration, error) {
	l, err := d.db.ListModuleConfiguration(ctx)
	if err != nil {
		return nil, dalerrs.TranslatePGError(err)
	}
	return l, nil
}
