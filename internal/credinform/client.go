package credinform

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"scoring_worker/internal/config"

	"go.uber.org/zap"
)

type Client struct {
	httpClient *http.Client
	config     *config.CredinformConfig
	logger     *zap.Logger
	accessKey  string
	username   string
	password   string
	apiVersion string
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessKey string `json:"accessKey"`
	Error     string `json:"error,omitempty"`
}

type SearchCompanyRequest struct {
	Language                string                  `json:"language"`
	SearchCompanyParameters SearchCompanyParameters `json:"searchCompanyParameters"`
}

type SearchCompanyParameters struct {
	TaxNumber string `json:"taxNumber"`
}

type SearchCompanyResponse struct {
	CompanyDataList []CompanyData `json:"companyDataList"`
	CommentRu       string        `json:"commentRu"`
	CommentEn       string        `json:"commentEn"`
}

type CompanyData struct {
	AddressVirtual      string         `json:"addressVirtual"`
	AddressLegal        string         `json:"addressLegal"`
	CodeRegionVirtual   string         `json:"codeRegionVirtual"`
	CodeRegionLegal     string         `json:"codeRegionLegal"`
	CompanyID           string         `json:"companyId"`
	CompanyName         string         `json:"companyName"`
	Country             string         `json:"country"`
	CountryCode         string         `json:"countryCode"`
	LegalForm           string         `json:"legalForm"`
	FoundationDateFloat FoundationDate `json:"foundationDateFloat"`
	StatisticalNumber   string         `json:"statisticalNumber"`
	RegistrationNumber  string         `json:"registrationNumber"`
	Status              string         `json:"status"`
	TaxNumber           string         `json:"taxNumber"`
	LastBalanceDate     string         `json:"lastBalanceDate"`
}

type FoundationDate struct {
	Year  int    `json:"year"`
	Month int    `json:"month"`
	Day   int    `json:"day"`
	Date  string `json:"date"`
}

type CompanyInformationRequest struct {
	CompanyID string                 `json:"companyId"`
	Language  string                 `json:"language"`
	Extra     map[string]interface{} `json:"-"`
}

type CompanyInformationResponse struct {
	Data  interface{} `json:"data"`
	Error string      `json:"error,omitempty"`
}

func NewClient(cfg *config.CredinformConfig, logger *zap.Logger) *Client {
	logger.Info("Creating Credinform client",
		zap.String("username", cfg.Username),
		zap.String("base_url", cfg.BaseURL),
		zap.Bool("password_set", cfg.Password != ""),
		zap.Int("timeout", cfg.Timeout))

	if cfg.Username == "" {
		logger.Error("Credinform username is empty")
		return nil
	}

	if cfg.Password == "" {
		logger.Error("Credinform password is empty")
		return nil
	}

	decodedPassword, err := base64.StdEncoding.DecodeString(cfg.Password)
	if err != nil {
		logger.Error("failed to decode credinform password from base64", zap.Error(err))
		return nil
	}
	return &Client{
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		config:     cfg,
		logger:     logger,
		username:   cfg.Username,
		password:   string(decodedPassword),
		apiVersion: "1.7",
	}
}

func (c *Client) Authenticate(ctx context.Context) error {
	authReq := AuthRequest{
		Username: c.username,
		Password: c.password,
	}

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		fmt.Sprintf("%s/api/Authorization/GetAccessKey", c.config.BaseURL),
		bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send auth request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read auth response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return fmt.Errorf("failed to unmarshal auth response: %w", err)
	}

	if authResp.Error != "" {
		return fmt.Errorf("auth error: %s", authResp.Error)
	}

	c.accessKey = authResp.AccessKey
	c.logger.Info("Successfully authenticated with Credinform API")
	return nil
}

func (c *Client) SearchCompany(ctx context.Context, inn string) (*CompanyData, error) {
	if c.accessKey == "" {
		if err := c.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	searchReq := SearchCompanyRequest{
		Language: "Russian",
		SearchCompanyParameters: SearchCompanyParameters{
			TaxNumber: inn,
		},
	}

	reqData, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	url := fmt.Sprintf("%s/api/Search/SearchCompany?apiVersion=%s",
		c.config.BaseURL, c.apiVersion)

	var result *CompanyData
	var lastErr error

	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			c.logger.Info("Retrying search request", zap.String("inn", inn), zap.Int("attempt", attempt))
			time.Sleep(time.Duration(c.config.RetryDelay) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create search request: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("accessKey", c.accessKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to send search request: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read search response: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			c.logger.Info("Token expired, re-authenticating")
			c.accessKey = ""
			if err := c.Authenticate(ctx); err != nil {
				lastErr = fmt.Errorf("failed to re-authenticate: %w", err)
				continue
			}
			req.Header.Set("accessKey", c.accessKey)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("search request failed with status %d: %s", resp.StatusCode, string(body))
			continue
		}

		var response SearchCompanyResponse
		if err := json.Unmarshal(body, &response); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal search response: %w", err)
			continue
		}

		if len(response.CompanyDataList) == 0 {
			return nil, fmt.Errorf("company not found")
		}

		result = &response.CompanyDataList[0]
		c.logger.Info("Successfully found company", zap.String("inn", inn), zap.String("company_id", result.CompanyID))
		break
	}

	if result == nil {
		return nil, fmt.Errorf("all retry attempts failed: %w", lastErr)
	}

	return result, nil
}

func structToMap(data interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}
	var m map[string]interface{}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &m)
	return m, err
}

func (c *Client) getCompanyData(ctx context.Context, method string, companyID string, params interface{}) ([]byte, error) {
	if c.accessKey == "" {
		if err := c.Authenticate(ctx); err != nil {
			return nil, fmt.Errorf("failed to authenticate for %s: %w", method, err)
		}
	}

	extra, err := structToMap(params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to map for %s: %w", method, err)
	}

	reqData, err := json.Marshal(CompanyInformationRequest{
		CompanyID: companyID,
		Language:  "Russian",
		Extra:     extra,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request for %s: %w", method, err)
	}

	url := fmt.Sprintf("%s/api/%s?apiVersion=%s", c.config.BaseURL, method, c.apiVersion)
	var responseBody []byte
	var allErrors []string

	for attempt := 0; attempt <= c.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			c.logger.Info("Retrying request", zap.String("method", method), zap.String("companyID", companyID), zap.Int("attempt", attempt))
			time.Sleep(time.Duration(c.config.RetryDelay) * time.Second)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqData))
		if err != nil {
			errMessage := fmt.Sprintf("failed to create request for %s: %v", method, err)
			allErrors = append(allErrors, errMessage)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("accessKey", c.accessKey)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errMessage := fmt.Sprintf("failed to send request for %s: %v", method, err)
			allErrors = append(allErrors, errMessage)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errMessage := fmt.Sprintf("failed to read response for %s: %v", method, err)
			allErrors = append(allErrors, errMessage)
			continue
		}

		if resp.StatusCode == http.StatusUnauthorized {
			c.logger.Info("Token expired, re-authenticating", zap.String("method", method))
			allErrors = append(allErrors, "token expired (401)")
			c.accessKey = ""
			if err := c.Authenticate(ctx); err != nil {
				errMessage := fmt.Sprintf("failed to re-authenticate for %s: %v", method, err)
				allErrors = append(allErrors, errMessage)
				continue
			}
			attempt-- // Retry the current request immediately after re-authenticating
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errMessage := fmt.Sprintf("request %s failed with status %d: %s", method, resp.StatusCode, string(body))
			allErrors = append(allErrors, errMessage)
			continue
		}

		responseBody = body
		return responseBody, nil // Success, exit loop
	}

	return nil, fmt.Errorf("all retry attempts failed for %s. Errors: [%s]", method, strings.Join(allErrors, "; "))
}
