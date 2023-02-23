package client

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	fieldCreatedAt = func() time.Time {
		t, _ := time.Parse(time.RFC3339, "2023-02-17T21:13:06.510Z")
		return t
	}
	fieldEmailSettings = map[string]any{
		"addresses": []any{
			map[string]any{"email": string("test1@host.com")},
			map[string]any{"email": string("test2@host.com")},
		},
	}
	fieldDesc = "testing..."
)

func TestSwoService_ReadNotification(t *testing.T) {
	ctx, client, server, _, teardown := setup()
	defer teardown()

	var settings any = fieldEmailSettings

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gqlInput, err := getGraphQLInput[__GetNotificationInput](r)
		if err != nil {
			t.Errorf("Swo.ReadNotification error: %v", err)
		}

		sendGraphQLResponse(t, w, GetNotificationResponse{
			User: GetNotificationUserAuthenticatedUser{
				CurrentOrganization: GetNotificationUserAuthenticatedUserCurrentOrganization{
					NotificationServiceConfiguration: ReadNotificationResult{
						Id:          gqlInput.ConfigurationId,
						Type:        gqlInput.ConfigurationType,
						Title:       "email test",
						Description: &fieldDesc,
						Settings:    &settings,
						CreatedAt:   fieldCreatedAt(),
						CreatedBy:   "140979956856880128",
					},
				},
			},
		})
	})

	got, err := client.NotificationsService().Read(ctx, "123", "email")
	if err != nil {
		t.Errorf("Swo.ReadNotification returned error: %v", err)
	}

	want := &ReadNotificationResult{
		Id:          "123",
		Title:       "email test",
		Description: &fieldDesc,
		Type:        "email",
		Settings:    &settings,
		CreatedAt:   fieldCreatedAt(),
		CreatedBy:   "140979956856880128",
	}

	if !testObjects(t, got, want) {
		t.Errorf("Swo.ReadNotification returned %+v, wanted %+v", got, want)
	}
}

func TestSwoService_CreateNotification(t *testing.T) {
	ctx, client, server, _, teardown := setup()
	defer teardown()

	var settings any = fieldEmailSettings

	requestInput := CreateNotificationInput{
		Title:       "email test",
		Description: &fieldDesc,
		Type:        "email",
		Settings:    settings,
	}

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gqlInput, err := getGraphQLInput[__CreateNotificationInput](r)
		if err != nil {
			t.Errorf("Swo.CreateNotification error: %v", err)
		}

		got := gqlInput.Configuration
		want := requestInput

		if !testObjects(t, got, want) {
			t.Errorf("Request got = %+v, want = %+v", got, want)
		}

		sendGraphQLResponse(t, w, CreateNotificationResponse{
			CreateNotificationServiceConfiguration: CreateNotificationCreateNotificationServiceConfigurationCreateNotificationServiceConfigurationResponse{
				Code:    "201",
				Success: true,
				Message: "",
				Configuration: &CreateNotificationResult{
					Id:          uuid.NewString(),
					Type:        got.Type,
					Title:       got.Title,
					Description: got.Description,
					Settings:    &got.Settings,
					CreatedAt:   fieldCreatedAt(),
					CreatedBy:   "140979956856880128",
				},
			},
		})
	})

	got, err := client.NotificationsService().Create(ctx, requestInput)
	if err != nil {
		t.Errorf("Swo.CreateNotification returned error: %v", err)
	}

	if got.Id == "" {
		t.Errorf("Swo.CreateNotification did not return an Id")
	}

	want := &CreateNotificationResult{
		Id:          got.Id,
		Title:       "email test",
		Description: &fieldDesc,
		Type:        "email",
		Settings:    &settings,
		CreatedAt:   got.CreatedAt,
		CreatedBy:   got.CreatedBy,
	}

	if !testObjects(t, got, want) {
		t.Errorf("Swo.CreateNotification returned %+v, want %+v", got, want)
	}
}

func TestSwoService_UpdateNotification(t *testing.T) {
	ctx, client, server, _, teardown := setup()
	defer teardown()

	var settings any = fieldEmailSettings
	nTitle := "email test"

	input := UpdateNotificationInput{
		Id:          "123",
		Title:       &nTitle,
		Description: &fieldDesc,
		Settings:    &settings,
	}

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gqlInput, err := getGraphQLInput[__UpdateNotificationInput](r)
		if err != nil {
			t.Errorf("Swo.UpdateNotification error: %v", err)
		}

		got := gqlInput.Configuration
		want := input

		if !testObjects(t, got, want) {
			t.Errorf("Request got = %+v, want = %+v", got, want)
		}

		sendGraphQLResponse(t, w, UpdateNotificationResponse{
			UpdateNotificationServiceConfiguration: &UpdateNotificationUpdateNotificationServiceConfigurationUpdateNotificationServiceConfigurationResponse{
				Code:    "201",
				Success: true,
				Message: "",
				Configuration: &UpdateNotificationUpdateNotificationServiceConfigurationUpdateNotificationServiceConfigurationResponseConfigurationNotificationService{
					Id:          got.Id,
					Title:       *got.Title,
					Description: got.Description,
					Settings:    got.Settings,
				},
			},
		})
	})

	err := client.NotificationsService().Update(ctx, input)
	if err != nil {
		t.Errorf("Swo.UpdateNotification returned error: %v", err)
	}
}

func TestSwoService_DeleteNotification(t *testing.T) {
	ctx, client, server, _, teardown := setup()
	defer teardown()

	input := DeleteNotificationServiceConfigurationInput{
		Id: "123",
	}

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gqlInput, err := getGraphQLInput[__DeleteNotificationInput](r)
		if err != nil {
			t.Errorf("Swo.DeleteNotification error: %v", err)
		}

		got := gqlInput.Input
		want := input

		if !testObjects(t, got, want) {
			t.Errorf("Swo.DeleteNotification: Request got = %+v, want %+v", got, want)
		}

		sendGraphQLResponse(t, w, DeleteNotificationResponse{
			DeleteNotificationServiceConfiguration: &DeleteNotificationDeleteNotificationServiceConfigurationDeleteNotificationServiceConfigurationResponse{
				Code:    "201",
				Success: true,
				Message: "",
			},
		})
	})

	err := client.NotificationsService().Delete(ctx, input.Id)
	if err != nil {
		t.Errorf("Swo.DeleteNotification returned error: %v", err)
	}
}

func TestSwoService_NotificationsServerErrors(t *testing.T) {
	ctx, client, server, _, teardown := setup()
	defer teardown()

	server.HandleFunc("/", httpErrorResponse)

	_, err := client.NotificationsService().Create(ctx, CreateNotificationInput{})
	if err == nil {
		t.Error("Swo.NotificationsServerErrors expected an error response")
	}
	_, err = client.NotificationsService().Read(ctx, "123", "email")
	if err == nil {
		t.Error("Swo.NotificationsServerErrors expected an error response")
	}
	err = client.NotificationsService().Update(ctx, UpdateNotificationInput{})
	if err == nil {
		t.Error("Swo.NotificationsServerErrors expected an error response")
	}
	err = client.NotificationsService().Delete(ctx, "123")
	if err == nil {
		t.Error("Swo.NotificationsServerErrors expected an error response")
	}
}

func TestNotification_Marshal(t *testing.T) {
	testJSONMarshal(t, &ReadNotificationResult{}, "{}")

	var settings any = fieldEmailSettings
	id := uuid.NewString()
	desc := "testing..."
	created := fieldCreatedAt()

	got := ReadNotificationResult{
		Id:          id,
		Title:       "email test",
		Description: &desc,
		Type:        "email",
		Settings:    &settings,
		CreatedAt:   created,
		CreatedBy:   "140979956856880128",
	}

	want := fmt.Sprintf(`{
		"id": "%s",
		"type": "email",
		"title": "email test",
		"settings": {
			"addresses": [
				{
					"email": "test1@host.com"
				},
				{
					"email": "test2@host.com"
				}
			]
		},
		"createdAt": "%s",
		"createdBy": "140979956856880128",
		"description": "testing..."
	}`, id, "2023-02-17T21:13:06.510Z")

	testJSONMarshal(t, got, want)
}
