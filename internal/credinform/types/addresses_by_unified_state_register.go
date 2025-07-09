package types

// --- Адреса по данным ЕГРЮЛ ---

type AddressesByUnifiedStateRegisterResponse = AddressesByUnifiedStateRegister

type AddressesByUnifiedStateRegister struct {
	AddressList []*AddressHistoryEgrul `json:"addressList,omitempty"`
}

type AddressHistoryEgrul struct {
	FromDate              *string `json:"fromDate,omitempty"`
	TillDate              *string `json:"tillDate,omitempty"`
	Address               *string `json:"address,omitempty"`
	IsAddressFalse        *bool   `json:"isAddressFalse,omitempty"`
	FalseAddressBeginDate *string `json:"falseAddressBeginDate,omitempty"`
	FalseAddressEndDate   *string `json:"falseAddressEndDate,omitempty"`
	IsMassAddressFTS      *bool   `json:"isMassAddressFTS,omitempty"`
	MassTillDate          *string `json:"massTillDate,omitempty"`
}
