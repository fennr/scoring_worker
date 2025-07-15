package credinform

import (
	"context"
	"encoding/json"
	"fmt"
	"scoring_worker/internal/credinform/types"
)

type ActivitiesParams struct{}

func (c *Client) GetActivities(ctx context.Context, companyID string, params ActivitiesParams) (*types.Activities, error) {
	body, err := c.getCompanyData(ctx, "CompanyInformation/Activities", companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	var response struct {
		Data types.Activities `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal activities response: %w", err)
	}

	return &response.Data, nil
}
