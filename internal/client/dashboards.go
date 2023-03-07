package client

import (
	"context"
	"fmt"
	"log"
)

type DashboardsService service

type CreateDashboardResult = createDashboardCreateDashboardCreateDashboardResponseDashboard
type CreateDashboardLayout = createDashboardCreateDashboardCreateDashboardResponseDashboardLayout
type CreateDashboardWidget = createDashboardCreateDashboardCreateDashboardResponseDashboardWidgetsWidget

type ReadDashboardResult = getDashboardByIdDashboardsDashboardQueriesByIdDashboard
type ReadDashboardLayout = getDashboardByIdDashboardsDashboardQueriesByIdDashboardLayout
type ReadDashboardWidget = getDashboardByIdDashboardsDashboardQueriesByIdDashboardWidgetsWidget

type UpdateDashboardResult = updateDashboardUpdateDashboardUpdateDashboardResponseDashboard
type UpdateDashboardLayout = updateDashboardUpdateDashboardUpdateDashboardResponseDashboardLayout
type UpdateDashboardWidget = updateDashboardUpdateDashboardUpdateDashboardResponseDashboardWidgetsWidget

type mutateHandler func() (any, error)

type DashboardsCommunicator interface {
	Create(context.Context, CreateDashboardInput) (*CreateDashboardResult, error)
	Read(context.Context, string) (*ReadDashboardResult, error)
	Update(context.Context, UpdateDashboardInput) (*UpdateDashboardResult, error)
	Delete(context.Context, string) error
}

func NewDashboardsService(c *Client) *DashboardsService {
	return &DashboardsService{c}
}

// Creates a new dashboard.
func (service *DashboardsService) Create(ctx context.Context, input CreateDashboardInput) (*CreateDashboardResult, error) {
	log.Printf("create dashboard request. name: %s", input.Name)

	resp, err := doMutate[createDashboardResponse](func() (any, error) {
		return createDashboard(ctx, service.client.gql, input)
	})

	if err != nil {
		return nil, err
	}

	dashboard := resp.CreateDashboard.Dashboard
	log.Printf("create dashboard success. id: %s", dashboard.Id)

	return dashboard, nil
}

// Returns the dashboard identified by the given Id.
func (service *DashboardsService) Read(ctx context.Context, id string) (*ReadDashboardResult, error) {
	log.Printf("read dashboard request. id: %s", id)

	resp, err := getDashboardById(ctx, service.client.gql, id)

	if err != nil {
		return nil, err
	}

	dashboard := resp.Dashboards.ById

	log.Printf("read dashboard success. name: %s", dashboard.Id)

	return dashboard, nil
}

// Updates the dashboard.
func (service *DashboardsService) Update(ctx context.Context, input UpdateDashboardInput) (*UpdateDashboardResult, error) {
	log.Printf("update dashboard request. id: %s", input.Id)

	resp, err := doMutate[updateDashboardResponse](func() (any, error) {
		return updateDashboard(ctx, service.client.gql, input)
	})

	if err != nil {
		return nil, err
	}

	log.Printf("update dashboard success. id: %s", input.Id)

	return resp.UpdateDashboard.Dashboard, nil
}

// Deletes the dashboard with the given id.
func (service *DashboardsService) Delete(ctx context.Context, id string) error {
	log.Printf("delete dashboard request. id: %s", id)

	_, err := doMutate[deleteDashboardResponse](func() (any, error) {
		return deleteDashboard(ctx, service.client.gql, DeleteDashboardInput{
			Id: id,
		})
	})

	if err != nil {
		return err
	}

	log.Printf("delete dashboard success. id: %s", id)

	return nil
}

func doMutate[T any](mutation mutateHandler) (*T, error) {
	resp, err := mutation()

	if err != nil {
		return nil, err
	}

	mutateError := func(mutationType string, code string, message string) error {
		return fmt.Errorf("%s dashboard returned a failure. code: %s message: %s",
			mutationType, code, message)
	}

	switch resp := resp.(type) {
	case *createDashboardResponse:
		if !resp.CreateDashboard.Success {
			err = mutateError("create", resp.CreateDashboard.Code, resp.CreateDashboard.Message)
		}
	case *updateDashboardResponse:
		if !resp.UpdateDashboard.Success {
			err = mutateError("update", resp.UpdateDashboard.Code, resp.UpdateDashboard.Message)
		}
	case *deleteDashboardResponse:
		if !resp.DeleteDashboard.Success {
			err = mutateError("delete", resp.DeleteDashboard.Code, resp.DeleteDashboard.Message)
		}
	default:
		return nil, fmt.Errorf("unexpected server response. resp: %+v", resp)
	}

	return resp.(*T), err
}
