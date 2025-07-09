package credinform

import (
	"context"
	"fmt"
	"scoring_worker/internal/credinform/types"

	"go.uber.org/zap"
)

type BasicInformationParams struct{}

func (c *Client) GetBasicInformation(ctx context.Context, companyID string, params BasicInformationParams) (types.BasicInformationResponse, error) {
	var result types.BasicInformationResponse
	resp, err := c.GetCompanyData(ctx, "BasicInformation", companyID, params)
	if err != nil {
		return result, err
	}

	// Логируем, что пришло от внешнего сервиса
	if c.logger != nil {
		c.logger.Info("GetCompanyData raw response (basic_information)",
			zap.String("company_id", companyID),
			zap.Any("resp_type", fmt.Sprintf("%T", resp)),
			zap.Any("resp_value", resp))
	}

	if err := DecodeToType(resp, &result); err != nil {
		return result, err
	}

	// Логируем, что получилось после DecodeToType
	if c.logger != nil {
		c.logger.Info("After DecodeToType (basic_information)",
			zap.String("company_id", companyID),
			zap.Any("result_type", fmt.Sprintf("%T", result)),
			zap.Any("result_value", result))
	}

	return result, nil
}
