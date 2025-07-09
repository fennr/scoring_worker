package types

// --- Адреса по данным Крединформ ---

type AddressesByCredinformResponse = AddressesByCredinform

type AddressesByCredinform struct {
	Addresses []*Address `json:"addresses,omitempty"`
}

type Address struct {
	AddressType        *string `json:"addressType,omitempty"`
	AddressTypeCode    *string `json:"addressTypeCode,omitempty"`
	CountryCode        *string `json:"countryCode,omitempty"`
	KladrCode          *string `json:"kladrCode,omitempty"`
	PostCode           *string `json:"postCode,omitempty"`
	Region             *Region `json:"region,omitempty"`
	District           *string `json:"district,omitempty"`
	City               *string `json:"city,omitempty"`
	Street             *string `json:"street,omitempty"`
	StreetShort        *string `json:"streetShort,omitempty"`
	HouseNumber        *string `json:"houseNumber,omitempty"`
	Housing            *string `json:"housing,omitempty"`
	Flat               *string `json:"flat,omitempty"`
	HouseFull          *string `json:"houseFull,omitempty"`
	IsOwned            *bool   `json:"isOwned,omitempty"`
	FalseAddressReason *string `json:"falseAddressReason,omitempty"`
	FalseAddressGRN    *string `json:"falseAddressGRN,omitempty"`
	FalseAddressDate   *string `json:"falseAddressDate,omitempty"`
	IsMassAddressFTS   *bool   `json:"isMassAddressFTS,omitempty"`
}

type Region struct {
	Name                *string          `json:"name,omitempty"`
	Code                *string          `json:"code,omitempty"`
	FederalDistrictCode *FederalDistrict `json:"federalDistrictCode,omitempty"`
}

type FederalDistrict struct {
	Name    *string `json:"name,omitempty"`
	Code    *int    `json:"code,omitempty"`
	Country *string `json:"country,omitempty"`
}
