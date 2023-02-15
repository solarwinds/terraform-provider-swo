package client

import (
	"context"
	"log"
)

type CreateNotificationInput = CreateNotificationServiceConfigurationInput
type CreateNotificationResult = CreateNotificationCreateNotificationServiceConfigurationCreateNotificationServiceConfigurationResponseConfigurationNotificationService
type ReadNotificationResult = GetNotificationUserAuthenticatedUserCurrentOrganizationNotificationServiceConfigurationNotificationService
type UpdateNotificationInput = UpdateNotificationServiceConfigurationInput

type NotificationsService struct {
	client *Client
}

type NotificationsCommunicator interface {
	Create(*CreateNotificationInput) (*CreateNotificationResult, error)
	Read(string, string) (*ReadNotificationResult, error)
	Update(*UpdateNotificationInput) error
	Delete(string) error
}

func NewNotificationsService(c *Client) *NotificationsService {
	return &NotificationsService{c}
}

// Creates a new notification.
func (as *NotificationsService) Create(input *CreateNotificationInput) (*CreateNotificationResult, error) {
	log.Printf("create notification request. title: %s", input.Title)

	ctx := context.Background()
	resp, err := CreateNotification(ctx, as.client.gql, *input)

	if err != nil {
		return nil, err
	}

	notification := resp.CreateNotificationServiceConfiguration.Configuration
	log.Printf("create notifications success. id: %s", notification.Id)

	return notification, nil
}

// Returns the notification identified by the given Id.
func (as *NotificationsService) Read(id string, notificationType string) (*ReadNotificationResult, error) {
	log.Printf("read notification request. id: %s", id)

	ctx := context.Background()
	resp, err := GetNotification(ctx, as.client.gql, id, notificationType)

	if err != nil {
		return nil, err
	}

	notification := resp.User.CurrentOrganization.NotificationServiceConfiguration

	log.Printf("read notification success. title: %s", notification.Title)

	return &notification, nil
}

// Updates the notification.
func (as *NotificationsService) Update(input *UpdateNotificationInput) error {
	log.Printf("update notification request. id: %s", input.Id)

	ctx := context.Background()
	_, err := UpdateNotification(ctx, as.client.gql, *input)

	if err != nil {
		return err
	}

	log.Printf("update notification success. id: %s", input.Id)

	return nil
}

// Deletes the notification with the given id.
func (as *NotificationsService) Delete(id string) error {
	log.Printf("delete notification request. id: %s", id)

	ctx := context.Background()
	_, err := DeleteNotification(ctx, as.client.gql, DeleteNotificationServiceConfigurationInput{
		Id: id,
	})

	if err != nil {
		return err
	}

	log.Printf("delete notification success. id: %s", id)

	return nil
}
