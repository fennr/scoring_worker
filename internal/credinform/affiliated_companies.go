package credinform

import (
	"context"
	"scoring_worker/internal/credinform/types"
)

type AffiliatedCompaniesParams struct {
	AffiliationTypes []string `json:"affiliationTypes"`
}

func (c *Client) GetAffiliatedCompanies(ctx context.Context, companyID string, params AffiliatedCompaniesParams) (types.AffiliatedCompaniesResponse, error) {
	var result types.AffiliatedCompaniesResponse
	resp, err := c.GetCompanyData(ctx, "AffiliatedCompanies", companyID, params)
	if err != nil {
		return result, err
	}
	if err := DecodeToType(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}
