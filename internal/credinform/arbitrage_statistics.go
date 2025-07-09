package credinform

import (
	"context"
	"scoring_worker/internal/credinform/types"
)

type ArbitrageStatisticsParams struct {
	LastCaseChangeDateRange struct {
		From string `json:"from"`
	} `json:"lastCaseChangeDateRange"`
	ArbitrageSideCommonType []string `json:"arbitrageSideCommonType"`
}

func (c *Client) GetArbitrageStatistics(ctx context.Context, companyID string, params ArbitrageStatisticsParams) (types.ArbitrageStatisticsResponse, error) {
	var result types.ArbitrageStatisticsResponse
	resp, err := c.GetCompanyData(ctx, "ArbitrageStatistics", companyID, params)
	if err != nil {
		return result, err
	}
	if err := DecodeToType(resp, &result); err != nil {
		return result, err
	}
	return result, nil
}
