package anki

import (
	"fmt"
)

type ModelNameService struct {
	client *Client
}

func NewModelNameService(client *Client) *ModelNameService {
	return &ModelNameService{client}
}

// ModelNames is a slice of model names.
type ModelNames []string

func (svc *ModelNameService) GetNames() (ModelNames, error) {
	var models ModelNames
	if err := svc.client.sendAndUnmarshal("modelNames", nil, &models); err != nil {
		return nil, fmt.Errorf("modelNames: %w", err)
	}
	return models, nil
}

// ModelNamesAndIds is a map of model names and their corresponding IDs.
type ModelNamesAndIds map[string]float64

func (svc *ModelNameService) GetNamesAndIds() (ModelNamesAndIds, error) {
	var models ModelNamesAndIds
	if err := svc.client.sendAndUnmarshal("modelNamesAndIds", nil, &models); err != nil {
		return nil, fmt.Errorf("modelNamesAndIds: %w", err)
	}
	return models, nil
}
