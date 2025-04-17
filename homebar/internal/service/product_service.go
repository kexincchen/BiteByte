package service

import (
	"context"
	"time"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository"
)

type ProductService struct {
	productRepo repository.ProductRepository
}

func NewProductService(pr repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: pr}
}

func (s *ProductService) Create(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	if err := s.productRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) GetByID(ctx context.Context, id uint) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, p *domain.Product) (*domain.Product, error) {
	p.UpdatedAt = time.Now()
	if err := s.productRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *ProductService) Delete(ctx context.Context, id uint) error {
	return s.productRepo.Delete(ctx, id)
}

func (s *ProductService) GetByMerchant(ctx context.Context, merchantID uint) ([]*domain.Product, error) {
	return s.productRepo.GetByMerchant(ctx, merchantID)
}

func (s *ProductService) GetAll(ctx context.Context) ([]*domain.Product, error) {
	return s.productRepo.GetAll(ctx)
}
