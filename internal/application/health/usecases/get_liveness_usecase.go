package usecases

import (
	"context"
	"time"

	"notification/internal/application/health/dtos"
)

type GetLivenessUseCase struct{}

func NewGetLivenessUseCase() *GetLivenessUseCase {
	return &GetLivenessUseCase{}
}

func (u *GetLivenessUseCase) Execute(ctx context.Context) (*dtos.LivenessResponse, error) {
	return &dtos.LivenessResponse{
		Status:    "OK",
		Timestamp: time.Now(),
	}, nil
}
