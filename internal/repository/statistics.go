package repository

import (
	"context"

	"github.com/usual2970/certimate/internal/app"
	"github.com/usual2970/certimate/internal/domain"
)

type StatisticsRepository struct{}

func NewStatisticsRepository() *StatisticsRepository {
	return &StatisticsRepository{}
}

func (r *StatisticsRepository) Get(ctx context.Context) (*domain.Statistics, error) {
	rs := &domain.Statistics{}
	// 所有证书
	certTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetApp().Dao().DB().NewQuery("select count(*) as total from certificate").One(&certTotal); err != nil {
		return nil, err
	}
	rs.CertificateTotal = certTotal.Total

	// 即将过期证书
	certExpireSoonTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetApp().Dao().DB().
		NewQuery("select count(*) as total from certificate where expireAt > datetime('now') and expireAt < datetime('now', '+20 days')").
		One(&certExpireSoonTotal); err != nil {
		return nil, err
	}
	rs.CertificateExpireSoon = certExpireSoonTotal.Total

	// 已过期证书
	certExpiredTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetApp().Dao().DB().
		NewQuery("select count(*) as total from certificate where expireAt < datetime('now')").
		One(&certExpiredTotal); err != nil {
		return nil, err
	}
	rs.CertificateExpired = certExpiredTotal.Total

	// 所有工作流
	workflowTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetApp().Dao().DB().NewQuery("select count(*) as total from workflow").One(&workflowTotal); err != nil {
		return nil, err
	}
	rs.WorkflowTotal = workflowTotal.Total

	// 已启用工作流
	workflowEnabledTotal := struct {
		Total int `db:"total"`
	}{}
	if err := app.GetApp().Dao().DB().NewQuery("select count(*) as total from workflow where enabled is TRUE").One(&workflowEnabledTotal); err != nil {
		return nil, err
	}
	rs.WorkflowEnabled = workflowEnabledTotal.Total
	rs.WorkflowDisabled = workflowTotal.Total - workflowEnabledTotal.Total

	return rs, nil
}
