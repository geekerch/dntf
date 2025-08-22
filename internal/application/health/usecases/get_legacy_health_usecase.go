package usecases

import (
	"context"
	"time"

	"notification/internal/application/health/dtos"
)

type GetLegacyHealthUseCase struct{}

func NewGetLegacyHealthUseCase() *GetLegacyHealthUseCase {
	return &GetLegacyHealthUseCase{}
}

func (u *GetLegacyHealthUseCase) Execute(ctx context.Context) (*dtos.LegacyHealthResponse, error) {
	return &dtos.LegacyHealthResponse{
		Status:    "OK",
		Message:   "Service is running",
		Timestamp: time.Now(),
	}, nil
}
