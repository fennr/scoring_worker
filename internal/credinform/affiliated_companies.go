package credinform

import (
	"context"
	"encoding/json"
	"fmt"
	"scoring_worker/internal/credinform/types"
)

type AffiliatedCompaniesParams struct {
	AffiliationTypes []string `json:"affiliationTypes"`
}

func (c *Client) GetAffiliatedCompanies(ctx context.Context, companyID string, params AffiliatedCompaniesParams) (*types.AffiliatedCompanies, error) {
	body, err := c.getCompanyData(ctx, "CompanyInformation/AffiliatedCompanies", companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get affiliated companies: %w", err)
	}

	var response struct {
		Data types.AffiliatedCompanies `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal affiliated companies response: %w", err)
	}

	return &response.Data, nil
}
