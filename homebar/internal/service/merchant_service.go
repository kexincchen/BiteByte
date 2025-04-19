package service

import (
	"context"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository"
)

type MerchantService struct{ repo repository.MerchantRepository }

func NewMerchantService(r repository.MerchantRepository) *MerchantService { return &MerchantService{r} }

func (s *MerchantService) List(ctx context.Context) ([]*domain.Merchant, error) {
	return s.repo.List(ctx)
}

func (s *MerchantService) GetByID(ctx context.Context, id uint) (*domain.Merchant, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MerchantService) GetByUsername(ctx context.Context, u string) (*domain.Merchant, error) {
	return s.repo.GetByUsername(ctx, u)
}

func (s *MerchantService) Create(ctx context.Context, m *domain.Merchant) error {
	now := time.Now()
	m.CreatedAt, m.UpdatedAt = now, now
	return s.repo.Create(ctx, m)
}

func (s *MerchantService) Update(ctx context.Context, m *domain.Merchant) error {
	m.UpdatedAt = time.Now()
	return s.repo.Update(ctx, m)
}

func (s *MerchantService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
