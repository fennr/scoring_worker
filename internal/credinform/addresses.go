package credinform

import (
	"context"
	"encoding/json"
	"fmt"
	"scoring_worker/internal/credinform/types"
)

type AddressesByCredinformParams struct{}

type AddressesByUnifiedStateRegisterParams struct{}

func (c *Client) GetAddressesByCredinform(ctx context.Context, companyID string, params AddressesByCredinformParams) (*types.AddressesByCredinform, error) {
	body, err := c.getCompanyData(ctx, "CompanyInformation/GetAddressesByCredinform", companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by credinform: %w", err)
	}

	var response struct {
		Data types.AddressesByCredinform `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal addresses by credinform response: %w", err)
	}

	return &response.Data, nil
}

func (c *Client) GetAddressesByUnifiedStateRegister(ctx context.Context, companyID string, params AddressesByUnifiedStateRegisterParams) (*types.AddressesByUnifiedStateRegister, error) {
	body, err := c.getCompanyData(ctx, "CompanyInformation/GetAddressesByUnifiedStateRegister", companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses by unified state register: %w", err)
	}

	var response struct {
		Data types.AddressesByUnifiedStateRegister `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal addresses by unified state register response: %w", err)
	}

	return &response.Data, nil
}
