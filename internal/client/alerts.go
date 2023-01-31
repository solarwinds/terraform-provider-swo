package client

import (
	"context"
	"log"
)

type AlertsService struct {
	client *Client
}

type AlertDefinitionCreate = CreateAlertDefinitionAlertMutationsCreateAlertDefinition
type AlertDefinitionUpdate = UpdateAlertDefinitionAlertMutationsUpdateAlertDefinition
type AlertDefinitionRead = GetAlertAlertQueriesAlertDefinitionsAlertDefinitionsResultAlertDefinitionsAlertDefinition

type AlertsCommunicator interface {
	Create(AlertDefinitionCreate) (*AlertDefinitionCreate, error)
	Read(string) (*AlertDefinitionRead, error)
	Update(AlertDefinitionUpdate) error
	Delete(string) error
}

func NewAlertsService(c *Client) *AlertsService {
	return &AlertsService{c}
}

// Creates a new alert with the given definition.
func (as *AlertsService) Create(a AlertDefinitionCreate) (*AlertDefinitionCreate, error) {
	log.Printf("Create alert request. Name: %s", a.Name)

	// a.Id = "0bc4710d-e3b0-4590-9c9b-e5e46d81d912"

	ctx := context.Background()
	resp, err := CreateAlertDefinition(ctx, as.client.gql)

	if err != nil {
		return nil, err
	}

	alertDef := resp.AlertMutations.CreateAlertDefinition
	log.Printf("Create alert success. Id: %s", alertDef.Id)

	return &alertDef, nil
}

// Returns the alert identified by the given Id.
func (as *AlertsService) Read(id string) (*AlertDefinitionRead, error) {
	log.Printf("Read alert request. Id: %s", id)

	ctx := context.Background()
	resp, err := GetAlert(ctx, as.client.gql)

	if err != nil {
		return nil, err
	}

	alertDef := resp.AlertQueries.AlertDefinitions.AlertDefinitions[0]

	log.Printf("Read alert success. Id: %s", id)

	return &alertDef, nil
}

// Updates the alert with the given id.
func (as *AlertsService) Update(a AlertDefinitionUpdate) error {
	log.Printf("Update alert request. Id: %s", a.Id)

	ctx := context.Background()
	_, err := UpdateAlertDefinition(ctx, as.client.gql)

	if err != nil {
		return err
	}

	log.Printf("Update alert success. Id: %s", a.Id)

	return nil
}

// Deletes the alert with the given id.
func (as *AlertsService) Delete(id string) error {
	log.Printf("Delete alert request. Id: %s", id)

	ctx := context.Background()
	_, err := DeleteAlertDefinition(ctx, as.client.gql)

	if err != nil {
		return err
	}

	log.Printf("Delete alert success. Id: %s", id)

	return nil
}
