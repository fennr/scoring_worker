package types

// --- Виды деятельности компании ---

type ActivitiesResponse = Activities

type Activities struct {
	KindOfActivityList       []*KindOfActivity      `json:"kindOfActivityList,omitempty"`
	TradeName                *string                `json:"tradeName,omitempty"`
	KindOfActivitiesRmspList []*KindOfActivityOkpd2 `json:"kindOfActivitiesRmspList,omitempty"`
}

type KindOfActivity struct {
	Industry *Industry `json:"industry,omitempty"`
	IsMain   *bool     `json:"isMain,omitempty"`
	IsActual *bool     `json:"isActual,omitempty"`
}

type KindOfActivityOkpd2 struct {
	Industry     *Industry `json:"industry,omitempty"`
	IsInnovation *bool     `json:"isInnovation,omitempty"`
}

type Industry struct {
	Name    *string `json:"name,omitempty"`
	Code    *string `json:"code,omitempty"`
	Group   *string `json:"group,omitempty"`
	Country *string `json:"country,omitempty"`
}
