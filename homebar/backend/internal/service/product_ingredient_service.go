package service

import (
	"context"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository/postgres"
)

type ProductIngredientService struct {
	productIngredientRepo *postgres.ProductIngredientRepository
}

func NewProductIngredientService(productIngredientRepo *postgres.ProductIngredientRepository) *ProductIngredientService {
	return &ProductIngredientService{
		productIngredientRepo: productIngredientRepo,
	}
}

func (s *ProductIngredientService) AddIngredientToProduct(ctx context.Context, productID, ingredientID int64, quantity float64) error {
	return s.productIngredientRepo.AddIngredientToProduct(ctx, productID, ingredientID, quantity)
}

func (s *ProductIngredientService) RemoveIngredientFromProduct(ctx context.Context, productID, ingredientID int64) error {
	return s.productIngredientRepo.RemoveIngredientFromProduct(ctx, productID, ingredientID)
}

func (s *ProductIngredientService) GetProductIngredients(ctx context.Context, productID int64) ([]*domain.ProductIngredient, error) {
	return s.productIngredientRepo.GetProductIngredients(ctx, productID)
} 