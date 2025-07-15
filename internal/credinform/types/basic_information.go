package types

// --- Основная информация о компании ---

type BasicInformationResponse = BasicInformation

type BasicInformation struct {
	BaseBaseCompany
	CaptionName                    *string      `json:"captionName,omitempty"`
	ShortName                      *string      `json:"shortName,omitempty"`
	InternationalName              *string      `json:"internationalName,omitempty"`
	Country                        *Country     `json:"country,omitempty"`
	Status                         *StatusType  `json:"status,omitempty"`
	StatisticalNumber              *string      `json:"statisticalNumber,omitempty"`
	LegalForm                      *LegalForm   `json:"legalForm,omitempty"`
	CompanyType                    *CompanyType `json:"companyType,omitempty"`
	PrimaryActivityCode            *string      `json:"primaryActivityCode,omitempty"`
	BusinessIdentificationNumber   *string      `json:"businessIdentificationNumber,omitempty"`
	CodesInClassifiersAndRegisters *Classifiers `json:"codesInClassifiersAndRegisters,omitempty"`
	BusinessEntity                 *string      `json:"businessEntity,omitempty"`
	IsLargestTaxpayer              *bool        `json:"isLargestTaxpayer,omitempty"`
	FoundationDateFloat            *FloatData   `json:"foundationDateFloat,omitempty"`
	LiquidationDateFloat           *FloatData   `json:"liquidationDateFloat,omitempty"`
	StatusDateFloat                *FloatData   `json:"statusDateFloat,omitempty"`
	TaxRegistrationReasonCode      *string      `json:"taxRegistrationReasonCode,omitempty"`
}

type BaseBaseCompany struct {
	CompanyID          string  `json:"companyId"`
	Name               *string `json:"name,omitempty"`
	TaxNumber          *string `json:"taxNumber,omitempty"`
	RegistrationNumber *string `json:"registrationNumber,omitempty"`
}

type Country struct {
	Name *string `json:"name,omitempty"`
	Code *string `json:"code,omitempty"`
}

type StatusType struct {
	Name     *string `json:"name,omitempty"`
	Code     *int    `json:"code,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
}

type LegalForm struct {
	Name      *string `json:"name,omitempty"`
	Code      *int    `json:"code,omitempty"`
	ShortName *string `json:"shortName,omitempty"`
	Branch    *bool   `json:"branch,omitempty"`
}

type CompanyType string

const (
	CompanyTypeOrdinaryCompany  CompanyType = "OrdinaryCompany"
	CompanyTypeBank             CompanyType = "Bank"
	CompanyTypeInsuranceCompany CompanyType = "InsuranceCompany"
	CompanyTypeIssuer           CompanyType = "Issuer"
)

type Classifiers struct {
	ClassifierList []*Classifier `json:"classifierList,omitempty"`
}

type Classifier struct {
	FullName  *string `json:"fullName,omitempty"`
	Code      *string `json:"code,omitempty"`
	ShortName *string `json:"shortName,omitempty"`
	Value     *string `json:"value,omitempty"`
	FromDate  *string `json:"fromDate,omitempty"`
	TillDate  *string `json:"tillDate,omitempty"`
}

type FloatData struct {
	Year  int     `json:"year"`
	Month *int    `json:"month,omitempty"`
	Day   *int    `json:"day,omitempty"`
	Date  *string `json:"date,omitempty"`
}
