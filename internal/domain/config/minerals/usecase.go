package minerals

import (
	"context"

	"github.com/gmhafiz/go8/internal/domain/config"
)

type UseCase interface {
	List(ctx context.Context) ([]*config.Mineral, error)
	GetByID(ctx context.Context, id int) (*config.Mineral, error)
	Create(ctx context.Context, req *config.CreateMineralRequest) (*config.Mineral, error)
	Update(ctx context.Context, id int, req *config.UpdateMineralRequest) (*config.Mineral, error)
	Delete(ctx context.Context, id int) error
}

type useCase struct {
	repo Repository
}

func NewUseCase(repo Repository) UseCase {
	return &useCase{repo: repo}
}

func (uc *useCase) List(ctx context.Context) ([]*config.Mineral, error) {
	return uc.repo.List(ctx)
}

func (uc *useCase) GetByID(ctx context.Context, id int) (*config.Mineral, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *useCase) Create(ctx context.Context, req *config.CreateMineralRequest) (*config.Mineral, error) {
	mineral := &config.Mineral{
		Name:        req.Name,
		Code:        req.Code,
		Description: req.Description,
	}

	err := uc.repo.Create(ctx, mineral)
	if err != nil {
		return nil, err
	}

	return mineral, nil
}

func (uc *useCase) Update(ctx context.Context, id int, req *config.UpdateMineralRequest) (*config.Mineral, error) {
	mineral, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		mineral.Name = req.Name
	}
	if req.Description != "" {
		mineral.Description = req.Description
	}
	if req.Active != nil {
		mineral.Active = *req.Active
	}

	err = uc.repo.Update(ctx, mineral)
	if err != nil {
		return nil, err
	}

	return mineral, nil
}

func (uc *useCase) Delete(ctx context.Context, id int) error {
	return uc.repo.Delete(ctx, id)
}
