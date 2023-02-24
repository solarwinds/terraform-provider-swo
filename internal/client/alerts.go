package client

import (
	"context"
	"log"
)

type AlertsService service

type AlertDefinitionCreate = CreateAlertDefinitionAlertMutationsCreateAlertDefinition
type AlertDefinitionUpdate = UpdateAlertDefinitionAlertMutationsUpdateAlertDefinition
type AlertDefinitionRead = GetAlertDefinitionsAlertQueriesAlertDefinitionsAlertDefinitionsResultAlertDefinitionsAlertDefinition

type AlertsCommunicator interface {
	Create(context.Context, AlertDefinitionInput) (*AlertDefinitionCreate, error)
	Read(context.Context, string) (*AlertDefinitionRead, error)
	Update(context.Context, string, AlertDefinitionInput) error
	Delete(context.Context, string) error
}

func NewAlertsService(c *Client) *AlertsService {
	return &AlertsService{c}
}

// Creates a new alert with the given definition.
func (as *AlertsService) Create(ctx context.Context, input AlertDefinitionInput) (*AlertDefinitionCreate, error) {
	log.Printf("Create alert request. Name: %s", input.Name)

	resp, err := CreateAlertDefinition(ctx, as.client.gql, input)

	if err != nil {
		return nil, err
	}

	alertDef := resp.AlertMutations.CreateAlertDefinition
	log.Printf("Create alert success. Id: %s", alertDef.Id)

	return alertDef, nil
}

// Returns the alert identified by the given Id.
func (as *AlertsService) Read(ctx context.Context, id string) (*AlertDefinitionRead, error) {
	log.Printf("Read alert request. Id: %s", id)

	filter := AlertFilterInput{
		Id: &id,
	}

	pagingFirst := 15
	paging := PagingInput{
		First: &pagingFirst,
	}

	sortDirection := SortDirectionDesc
	sortBy := SortInput{
		Sorts: []SortItemInput{
			{
				PropertyName: "id",
				Direction:    &sortDirection,
			},
		},
	}

	resp, err := GetAlertDefinitions(ctx, as.client.gql, filter, &paging, &sortBy)

	if err != nil {
		return nil, err
	}

	alertDef := resp.AlertQueries.AlertDefinitions.AlertDefinitions[0]

	log.Printf("Read alert success. Id: %s", id)

	return &alertDef, nil
}

// Updates the alert with the given id.
func (as *AlertsService) Update(ctx context.Context, id string, input AlertDefinitionInput) error {
	log.Printf("Update alert request. Id: %s", id)

	_, err := UpdateAlertDefinition(ctx, as.client.gql, input, id)

	if err != nil {
		return err
	}

	log.Printf("Update alert success. Id: %s", id)

	return nil
}

// Deletes the alert with the given id.
func (as *AlertsService) Delete(ctx context.Context, id string) error {
	log.Printf("Delete alert request. Id: %s", id)

	_, err := DeleteAlertDefinition(ctx, as.client.gql, id)

	if err != nil {
		return err
	}

	log.Printf("Delete alert success. Id: %s", id)

	return nil
}
