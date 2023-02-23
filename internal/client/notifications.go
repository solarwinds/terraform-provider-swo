package client

import (
	"context"
	"log"
)

type CreateNotificationInput = CreateNotificationServiceConfigurationInput
type CreateNotificationResult = CreateNotificationCreateNotificationServiceConfigurationCreateNotificationServiceConfigurationResponseConfigurationNotificationService
type ReadNotificationResult = GetNotificationUserAuthenticatedUserCurrentOrganizationNotificationServiceConfigurationNotificationService
type UpdateNotificationInput = UpdateNotificationServiceConfigurationInput

type NotificationsService service

type NotificationsCommunicator interface {
	Create(context.Context, CreateNotificationInput) (*CreateNotificationResult, error)
	Read(context.Context, string, string) (*ReadNotificationResult, error)
	Update(context.Context, UpdateNotificationInput) error
	Delete(context.Context, string) error
}

func NewNotificationsService(c *Client) *NotificationsService {
	return &NotificationsService{c}
}

// Creates a new notification.
func (service *NotificationsService) Create(ctx context.Context, input CreateNotificationInput) (*CreateNotificationResult, error) {
	log.Printf("create notification request. title: %s", input.Title)

	resp, err := CreateNotification(ctx, service.client.gql, input)

	if err != nil {
		return nil, err
	}

	notification := resp.CreateNotificationServiceConfiguration.Configuration
	log.Printf("create notifications success. id: %s", notification.Id)

	return notification, nil
}

// Returns the notification identified by the given Id.
func (service *NotificationsService) Read(ctx context.Context, id string, notificationType string) (*ReadNotificationResult, error) {
	log.Printf("read notification request. id: %s", id)

	resp, err := GetNotification(ctx, service.client.gql, id, notificationType)

	if err != nil {
		return nil, err
	}

	notification := resp.User.CurrentOrganization.NotificationServiceConfiguration

	log.Printf("read notification success. title: %s", notification.Title)

	return &notification, nil
}

// Updates the notification.
func (service *NotificationsService) Update(ctx context.Context, input UpdateNotificationInput) error {
	log.Printf("update notification request. id: %s", input.Id)

	_, err := UpdateNotification(ctx, service.client.gql, input)

	if err != nil {
		return err
	}

	log.Printf("update notification success. id: %s", input.Id)

	return nil
}

// Deletes the notification with the given id.
func (service *NotificationsService) Delete(ctx context.Context, id string) error {
	log.Printf("delete notification request. id: %s", id)

	_, err := DeleteNotification(ctx, service.client.gql, DeleteNotificationServiceConfigurationInput{
		Id: id,
	})

	if err != nil {
		return err
	}

	log.Printf("delete notification success. id: %s", id)

	return nil
}
