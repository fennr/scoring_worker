package credinform

import (
	"context"
	"scoring_worker/internal/credinform/types"
)

type ActivitiesParams struct{}

func (c *Client) GetActivities(ctx context.Context, companyID string, params ActivitiesParams) (types.ActivitiesResponse, error) {
	var result types.ActivitiesResponse
	resp, err := c.GetCompanyData(ctx, "Activities", companyID, params)
	if err != nil {
		return result, err
	}
	if err := DecodeToType(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}
