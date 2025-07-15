package credinform

import (
	"context"
	"encoding/json"
	"fmt"
	"scoring_worker/internal/credinform/types"
)

type BasicInformationParams struct{}

func (c *Client) GetBasicInformation(ctx context.Context, companyID string, params BasicInformationParams) (*types.BasicInformation, error) {
	body, err := c.getCompanyData(ctx, "CompanyInformation/GetBasicInformation", companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic information: %w", err)
	}

	var response struct {
		Data types.BasicInformation `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal basic information response: %w", err)
	}

	return &response.Data, nil
}
