package credinform

import (
	"context"
	"encoding/json"
	"fmt"
	"scoring_worker/internal/credinform/types"
)

type ArbitrageStatisticsParams struct {
	LastCaseChangeDateRange struct {
		From string `json:"from"`
	} `json:"lastCaseChangeDateRange"`
	ArbitrageSideCommonType []string `json:"arbitrageSideCommonType"`
}

func (c *Client) GetArbitrageStatistics(ctx context.Context, companyID string, params ArbitrageStatisticsParams) (*types.ArbitrageStatistics, error) {
	body, err := c.getCompanyData(ctx, "CompanyInformation/ArbitrageStatistics", companyID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get arbitrage statistics: %w", err)
	}

	var response struct {
		Data types.ArbitrageStatistics `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal arbitrage statistics response: %w", err)
	}

	return &response.Data, nil
}
