package usecases

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"notification/internal/application/channel/dtos"
	"notification/internal/domain/channel"
	"notification/internal/domain/services"
	"notification/pkg/config"
)

// DeleteChannelUseCase is the use case for deleting a channel.
type DeleteChannelUseCase struct {
	channelRepo channel.ChannelRepository
	validator   *services.ChannelValidator
	config      *config.Config
}

// NewDeleteChannelUseCase creates a use case instance.
func NewDeleteChannelUseCase(
	channelRepo channel.ChannelRepository,
	validator *services.ChannelValidator,
	config *config.Config,
) *DeleteChannelUseCase {
	return &DeleteChannelUseCase{
		channelRepo: channelRepo,
		validator:   validator,
		config:      config,
	}
}

// Execute executes the delete channel operation.
func (uc *DeleteChannelUseCase) Execute(ctx context.Context, channelID string) (*dtos.DeleteChannelResponse, error) {
	// 1. Validate input parameters
	if channelID == "" {
		return nil, fmt.Errorf("channel ID is required")
	}

	// 2. Convert to domain object
	id, err := channel.NewChannelIDFromString(channelID)
	if err != nil {
		return nil, fmt.Errorf("invalid channel ID: %w", err)
	}

	// 3. Business validation
	if err := uc.validator.ValidateChannelDeletion(ctx, id); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 4. Query the channel
	ch, err := uc.channelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("channel not found: %w", err)
	}

	// 5. Forward to legacy system
	if err := uc.forwardDeleteToLegacySystem(ctx, ch.ID().String()); err != nil {
		return nil, fmt.Errorf("failed to forward delete to legacy system: %w", err)
	}

	// 6. Perform soft deletion
	if err := ch.Delete(); err != nil {
		return nil, fmt.Errorf("failed to delete channel: %w", err)
	}

	// 7. Persist
	if err := uc.channelRepo.Update(ctx, ch); err != nil {
		return nil, fmt.Errorf("failed to save channel deletion: %w", err)
	}

	// 8. Convert to response DTO
	response := &dtos.DeleteChannelResponse{
		ChannelID: ch.ID().String(),
		Deleted:   true,
		DeletedAt: *ch.Timestamps().DeletedAt,
	}

	return response, nil
}

// forwardDeleteToLegacySystem forwards the delete request to the legacy system
func (uc *DeleteChannelUseCase) forwardDeleteToLegacySystem(ctx context.Context, groupID string) error {
	legacyURL := uc.config.LegacySystem.URL + "/api/v2.0/Groups"
	bearerToken := uc.config.LegacySystem.Token

	// 1. Construct the request body for the legacy system (array of group IDs)
	reqBody := []string{groupID}

	// 2. Marshal the request body to JSON
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal legacy request body: %w", err)
	}

	// 3. Create and send the HTTP DELETE request
	req, err := http.NewRequestWithContext(ctx, "DELETE", legacyURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create legacy http request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to legacy system: %w", err)
	}
	defer resp.Body.Close()

	// 4. Check response status
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("legacy system returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}