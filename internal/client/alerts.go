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
type AlertDefinitionRead = GetAlertDefinitionsAlertQueriesAlertDefinitionsAlertDefinitionsResultAlertDefinitionsAlertDefinition

type AlertsCommunicator interface {
	Create(AlertDefinitionInput) (*AlertDefinitionCreate, error)
	Read(string) (*AlertDefinitionRead, error)
	Update(string, AlertDefinitionInput) error
	Delete(string) error
}

func NewAlertsService(c *Client) *AlertsService {
	return &AlertsService{c}
}

// Creates a new alert with the given definition.
func (as *AlertsService) Create(input AlertDefinitionInput) (*AlertDefinitionCreate, error) {
	log.Printf("Create alert request. Name: %s", input.Name)

	ctx := context.Background()
	resp, err := CreateAlertDefinition(ctx, as.client.gql, input)

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
	filter := AlertFilterInput{
		Id: id,
	}
	paging := PagingInput{}
	sortBy := SortInput{}

	resp, err := GetAlertDefinitions(ctx, as.client.gql, filter, paging, sortBy)

	if err != nil {
		return nil, err
	}

	alertDef := resp.AlertQueries.AlertDefinitions.AlertDefinitions[0]

	log.Printf("Read alert success. Id: %s", id)

	return &alertDef, nil
}

// Updates the alert with the given id.
func (as *AlertsService) Update(id string, input AlertDefinitionInput) error {
	log.Printf("Update alert request. Id: %s", id)

	ctx := context.Background()
	_, err := UpdateAlertDefinition(ctx, as.client.gql, input, id)

	if err != nil {
		return err
	}

	log.Printf("Update alert success. Id: %s", id)

	return nil
}

// Deletes the alert with the given id.
func (as *AlertsService) Delete(id string) error {
	log.Printf("Delete alert request. Id: %s", id)

	ctx := context.Background()
	_, err := DeleteAlertDefinition(ctx, as.client.gql, id)

	if err != nil {
		return err
	}

	log.Printf("Delete alert success. Id: %s", id)

	return nil
}
