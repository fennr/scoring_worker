package types

// --- Аффилированные компании ---

type AffiliatedCompaniesResponse = AffiliatedCompanies

type AffiliatedCompanies struct {
	AffUnderAdministrationList              []*AffiliationUnderAdministration              `json:"affUnderAdministrationList,omitempty"`
	AffToManagersSharesByNaturalPersonsList []*AffiliationToManagersSharesByNaturalPersons `json:"affToManagersSharesByNaturalPersonsList,omitempty"`
	AffToManagingCompanyList                []*AffiliationToManagingCompany                `json:"affToManagingCompanyList,omitempty"`
	AffToShareholdersLegalPersonsList       []*AffiliationToShareholdersLegalPersons       `json:"affToShareholdersLegalPersonsList,omitempty"`
	AffToLiquidatorOrBankruptcyAdminList    []*AffiliationToLiquidatorOrBankruptcyAdmin    `json:"affToLiquidatorOrBankruptcyAdminList,omitempty"`
}

type AffiliationUnderAdministration struct {
	BaseCompany
	AffiliatedNaturalPersonList []*AffiliatedNaturalPerson `json:"affiliatedNaturalPersonList,omitempty"`
}

type AffiliationToManagersSharesByNaturalPersons struct {
	BaseCompany
	AffiliatedNaturalPersonList []*AffiliatedNaturalPerson `json:"affiliatedNaturalPersonList,omitempty"`
}

type AffiliationToManagingCompany struct {
	BaseCompany
	AffiliatedCompanyList []*AffiliatedCompany `json:"affiliatedCompanyList,omitempty"`
}

type AffiliationToShareholdersLegalPersons struct {
	BaseCompany
	AffiliatedCompanyList     []*AffiliatedCompany     `json:"affiliatedCompanyList,omitempty"`
	AffiliatedLegalEntityList []*AffiliatedLegalEntity `json:"affiliatedLegalEntityList,omitempty"`
}

type AffiliationToLiquidatorOrBankruptcyAdmin struct {
	BaseCompany
	AffiliatedNaturalPersonList []*AffiliatedNaturalPerson `json:"affiliatedNaturalPersonList,omitempty"`
}

type AffiliatedCompany struct {
	BaseCompany
	ContributionPercent *float64 `json:"contributionPercent,omitempty"`
}

type AffiliatedLegalEntity struct {
	LegalEntity
	ContributionPercent *float64 `json:"contributionPercent,omitempty"`
}

type AffiliatedNaturalPerson struct {
	NaturalPerson
	ManagerPositions    []*ManagementPosition `json:"managerPositions,omitempty"`
	ContributionPercent *float64              `json:"contributionPercent,omitempty"`
}

type LegalEntity struct {
	Name               *string  `json:"name,omitempty"`
	Country            *Country `json:"country,omitempty"`
	Address            *string  `json:"address,omitempty"`
	PhoneNumber        *string  `json:"phoneNumber,omitempty"`
	FaxNumber          *string  `json:"faxNumber,omitempty"`
	TaxNumber          *string  `json:"taxNumber,omitempty"`
	RegistrationNumber *string  `json:"registrationNumber,omitempty"`
	StatisticalNumber  *string  `json:"statisticalNumber,omitempty"`
}

type NaturalPerson struct {
	PhysicalId   *string `json:"physicalId,omitempty"`
	TaxNumber    *string `json:"taxNumber,omitempty"`
	FirstNameRu  *string `json:"firstNameRu,omitempty"`
	FirstNameEn  *string `json:"firstNameEn,omitempty"`
	PatronymicRu *string `json:"patronymicRu,omitempty"`
	PatronymicEn *string `json:"patronymicEn,omitempty"`
	LastNameRu   *string `json:"lastNameRu,omitempty"`
	LastNameEn   *string `json:"lastNameEn,omitempty"`
	Nationality  *string `json:"nationality,omitempty"`
	DateOfBirth  *string `json:"dateOfBirth,omitempty"`
}

type ManagementPosition struct {
	BaseDictionaryInfo
	ManagementPositionGroup *ManagementPositionGroup `json:"managementPositionGroup,omitempty"`
}

type ManagementPositionGroup struct {
	BaseDictionaryInfo
	PositionLevel *int `json:"positionLevel,omitempty"`
}

type BaseDictionaryInfo struct {
	Name *string `json:"name,omitempty"`
	Code *int    `json:"code,omitempty"`
}
