package service

import (
	"context"

	"github.com/kexincchen/homebar/internal/domain"
	"github.com/kexincchen/homebar/internal/repository/postgres"
)

type IngredientService struct {
	ingredientRepo *postgres.IngredientRepository
}

func NewIngredientService(ingredientRepo *postgres.IngredientRepository) *IngredientService {
	return &IngredientService{
		ingredientRepo: ingredientRepo,
	}
}

func (s *IngredientService) CreateIngredient(ctx context.Context, ingredient *domain.Ingredient) (*domain.Ingredient, error) {
	return s.ingredientRepo.Create(ctx, ingredient)
}

func (s *IngredientService) GetIngredientByID(ctx context.Context, id int64) (*domain.Ingredient, error) {
	return s.ingredientRepo.GetByID(ctx, id)
}

func (s *IngredientService) GetIngredientsByMerchant(ctx context.Context, merchantID int64) ([]*domain.Ingredient, error) {
	return s.ingredientRepo.GetByMerchant(ctx, merchantID)
}

func (s *IngredientService) UpdateIngredient(ctx context.Context, ingredient *domain.Ingredient) error {
	return s.ingredientRepo.Update(ctx, ingredient)
}

func (s *IngredientService) DeleteIngredient(ctx context.Context, id int64) error {
	return s.ingredientRepo.Delete(ctx, id)
}

func (s *IngredientService) GetInventorySummary(ctx context.Context, merchantID int64) (map[string]interface{}, error) {
	return s.ingredientRepo.GetInventorySummary(ctx, merchantID)
}

func (s *IngredientService) HasSufficientInventoryForOrder(ctx context.Context, orderItems []*domain.OrderItem) (bool, error) {
	return s.ingredientRepo.LockInventoryForOrder(ctx, orderItems)
} 