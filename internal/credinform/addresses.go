package credinform

import (
	"context"
	"scoring_worker/internal/credinform/types"
)

type AddressesByCredinformParams struct{}
type AddressesByUnifiedStateRegisterParams struct{}

func (c *Client) GetAddressesByCredinform(ctx context.Context, companyID string, params AddressesByCredinformParams) (types.AddressesByCredinformResponse, error) {
	var result types.AddressesByCredinformResponse
	resp, err := c.GetCompanyData(ctx, "AddressesByCredinform", companyID, params)
	if err != nil {
		return result, err
	}
	if err := DecodeToType(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) GetAddressesByUnifiedStateRegister(ctx context.Context, companyID string, params AddressesByUnifiedStateRegisterParams) (types.AddressesByUnifiedStateRegisterResponse, error) {
	var result types.AddressesByUnifiedStateRegisterResponse
	resp, err := c.GetCompanyData(ctx, "AddressesByUnifiedStateRegister", companyID, params)
	if err != nil {
		return result, err
	}
	if err := DecodeToType(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}
