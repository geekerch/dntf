package usecases

import (
	"context"
	"fmt"

	"notification/internal/application/message/dtos"
	"notification/internal/domain/message"
)

// ListMessagesUseCase is the use case for listing messages.
type ListMessagesUseCase struct {
	messageRepo message.MessageRepository
}

// NewListMessagesUseCase creates a use case instance.
func NewListMessagesUseCase(
	messageRepo message.MessageRepository,
) *ListMessagesUseCase {
	return &ListMessagesUseCase{
		messageRepo: messageRepo,
	}
}

// Execute executes the list messages operation.
func (uc *ListMessagesUseCase) Execute(ctx context.Context, request *dtos.ListMessagesRequest) (*dtos.ListMessagesResponse, error) {
	// 1. Validate input parameters
	if err := uc.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// 2. For now, return empty result since repository doesn't support listing
	// TODO: Extend MessageRepository interface to support listing with filters
	response := uc.createEmptyResponse(request)
	return response, nil
}

// validateRequest validates the request parameters.
func (uc *ListMessagesUseCase) validateRequest(request *dtos.ListMessagesRequest) error {
	if request == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate pagination parameters
	if request.MaxResultCount < 0 {
		return fmt.Errorf("maxResultCount cannot be negative")
	}
	if request.SkipCount < 0 {
		return fmt.Errorf("skipCount cannot be negative")
	}
	if request.MaxResultCount > 1000 {
		return fmt.Errorf("maxResultCount cannot exceed 1000")
	}

	// Set default pagination if not provided
	if request.MaxResultCount == 0 {
		request.MaxResultCount = 10
	}

	return nil
}

// createEmptyResponse creates an empty response for now.
func (uc *ListMessagesUseCase) createEmptyResponse(request *dtos.ListMessagesRequest) *dtos.ListMessagesResponse {
	return &dtos.ListMessagesResponse{
		Items:          []*dtos.MessageResponse{},
		SkipCount:      request.SkipCount,
		MaxResultCount: request.MaxResultCount,
		TotalCount:     0,
		HasMore:        false,
	}
}