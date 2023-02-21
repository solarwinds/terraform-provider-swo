package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Khan/genqlient/graphql"
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
	client, server, _, teardown := setup()
	defer teardown()

	var settings any = fieldEmailSettings

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var request graphql.Request

		// Decode the graphql request object from the body.
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			t.Errorf("Swo.ReadNotification error: %v", err)
		}

		// Decode the variables field of the request to obtain the input variables.
		vars, err := ConvertObject[__GetNotificationInput](request.Variables)
		if err != nil {
			t.Errorf("Swo.ReadNotification error: %v", err)
		}

		response := graphql.Response{
			Data: GetNotificationResponse{
				User: GetNotificationUserAuthenticatedUser{
					CurrentOrganization: GetNotificationUserAuthenticatedUserCurrentOrganization{
						NotificationServiceConfiguration: ReadNotificationResult{
							Id:          vars.ConfigurationId,
							Type:        vars.ConfigurationType,
							Title:       "email test",
							Description: &fieldDesc,
							Settings:    &settings,
							CreatedAt:   fieldCreatedAt(),
							CreatedBy:   "140979956856880128",
						},
					},
				},
			},
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("Swo.ReadNotification error: %v", err)
		}
	})

	got, err := client.NotificationsService().Read("123", "email")
	if err != nil {
		t.Errorf("Swo.ReadNotification returned error: %v", err)
	}

	desc := "testing..."

	want := &ReadNotificationResult{
		Id:          "123",
		Title:       "email test",
		Description: &desc,
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
	client, server, _, teardown := setup()
	defer teardown()

	var settings any = fieldEmailSettings

	requestInput := CreateNotificationInput{
		Title:       "email test",
		Description: &fieldDesc,
		Type:        "email",
		Settings:    settings,
	}

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		request := new(graphql.Request)
		err := json.NewDecoder(r.Body).Decode(request)
		if err != nil {
			t.Errorf("Swo.CreateNotification error: %v", err)
		}

		vars, err := ConvertObject[__CreateNotificationInput](request.Variables)
		if err != nil {
			t.Errorf("Swo.CreateNotification error: %v", err)
		}

		inputConfig := vars.Configuration

		if !testObjects(t, inputConfig, requestInput) {
			t.Errorf("Request input = %+v, want %+v", inputConfig, requestInput)
		}

		response := graphql.Response{
			Data: CreateNotificationResponse{
				CreateNotificationServiceConfiguration: CreateNotificationCreateNotificationServiceConfigurationCreateNotificationServiceConfigurationResponse{
					Code:    "201",
					Success: true,
					Message: "",
					Configuration: &CreateNotificationResult{
						Id:          uuid.NewString(),
						Type:        inputConfig.Type,
						Title:       inputConfig.Title,
						Description: inputConfig.Description,
						Settings:    &inputConfig.Settings,
						CreatedAt:   fieldCreatedAt(),
						CreatedBy:   "140979956856880128",
					},
				},
			},
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("Swo.CreateNotification error: %v", err)
		}
	})

	got, err := client.NotificationsService().Create(&requestInput)
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
	client, server, _, teardown := setup()
	defer teardown()

	var settings any = fieldEmailSettings
	nTitle := "email test"

	requestInput := UpdateNotificationInput{
		Id:          "123",
		Title:       &nTitle,
		Description: &fieldDesc,
		Settings:    &settings,
	}

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		request := new(graphql.Request)
		err := json.NewDecoder(r.Body).Decode(request)
		if err != nil {
			t.Errorf("Swo.UpdateNotification error: %v", err)
		}

		vars, err := ConvertObject[__UpdateNotificationInput](request.Variables)
		if err != nil {
			t.Errorf("Swo.UpdateNotification error: %v", err)
		}

		inputConfig := vars.Configuration

		if !testObjects(t, inputConfig, requestInput) {
			t.Errorf("Request input = %+v, want %+v", inputConfig, requestInput)
		}

		response := graphql.Response{
			Data: UpdateNotificationResponse{
				UpdateNotificationServiceConfiguration: &UpdateNotificationUpdateNotificationServiceConfigurationUpdateNotificationServiceConfigurationResponse{
					Code:    "201",
					Success: true,
					Message: "",
					Configuration: &UpdateNotificationUpdateNotificationServiceConfigurationUpdateNotificationServiceConfigurationResponseConfigurationNotificationService{
						Id:          inputConfig.Id,
						Title:       *inputConfig.Title,
						Description: inputConfig.Description,
						Settings:    inputConfig.Settings,
					},
				},
			},
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("Swo.UpdateNotification error: %v", err)
		}
	})

	err := client.NotificationsService().Update(&requestInput)
	if err != nil {
		t.Errorf("Swo.UpdateNotification returned error: %v", err)
	}
}

func TestSwoService_DeleteNotification(t *testing.T) {
	client, server, _, teardown := setup()
	defer teardown()

	input := DeleteNotificationServiceConfigurationInput{
		Id: "123",
	}

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		request := new(graphql.Request)
		err := json.NewDecoder(r.Body).Decode(request)
		if err != nil {
			t.Errorf("Swo.DeleteNotification error: %v", err)
		}

		varsBytes, err := json.Marshal(request.Variables)
		if err != nil {
			t.Errorf("Swo.DeleteNotification error: %v", err)
		}

		var requestInput __DeleteNotificationInput
		err = json.Unmarshal(varsBytes, &requestInput)
		if err != nil {
			t.Errorf("Swo.DeleteNotification error: %v", err)
		}

		config := requestInput.Input

		if !testObjects(t, config, input) {
			t.Errorf("Swo.DeleteNotification: Request body = %+v, want %+v", config, input)
		}

		response := graphql.Response{
			Data: DeleteNotificationResponse{
				DeleteNotificationServiceConfiguration: &DeleteNotificationDeleteNotificationServiceConfigurationDeleteNotificationServiceConfigurationResponse{
					Code:    "201",
					Success: true,
					Message: "",
				},
			},
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			t.Errorf("Swo.DeleteNotification error: %v", err)
		}
	})

	err := client.NotificationsService().Delete(input.Id)
	if err != nil {
		t.Errorf("Swo.DeleteNotification returned error: %v", err)
	}
}

func TestSwoService_ServerErrors(t *testing.T) {
	client, server, _, teardown := setup()
	defer teardown()

	server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := client.NotificationsService().Create(&CreateNotificationInput{})
	if err == nil {
		t.Error("Swo.CreateNotificationError expected an error response")
	}
	_, err = client.NotificationsService().Read("123", "email")
	if err == nil {
		t.Error("Swo.ReadNotificationError expected an error response")
	}
	err = client.NotificationsService().Update(&UpdateNotificationInput{})
	if err == nil {
		t.Error("Swo.UpdateNotificationError expected an error response")
	}
	err = client.NotificationsService().Delete("123")
	if err == nil {
		t.Error("Swo.DeleteNotificationError expected an error response")
	}
}

func TestNotification_Marshal(t *testing.T) {
	testJSONMarshal(t, &ReadNotificationResult{}, "{}")

	var settings any = fieldEmailSettings
	id := uuid.NewString()
	desc := "testing..."
	created := fieldCreatedAt()

	got := &ReadNotificationResult{
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
