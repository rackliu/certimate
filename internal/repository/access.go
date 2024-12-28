package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/usual2970/certimate/internal/app"
	"github.com/usual2970/certimate/internal/domain"
)

type AccessRepository struct{}

func NewAccessRepository() *AccessRepository {
	return &AccessRepository{}
}

func (a *AccessRepository) GetById(ctx context.Context, id string) (*domain.Access, error) {
	record, err := app.GetApp().Dao().FindRecordById("access", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}

	rs := &domain.Access{
		Meta: domain.Meta{
			Id:        record.GetId(),
			CreatedAt: record.GetTime("created"),
			UpdatedAt: record.GetTime("updated"),
		},
		Name:       record.GetString("name"),
		ConfigType: record.GetString("configType"),
		Config:     record.GetString("config"),
		Usage:      record.GetString("usage"),
	}
	return rs, nil
}
