package client

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/solarwindscloud/terraform-provider-swo/internal/client/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	createResponseJson = `{"data":{"alertMutations":{"createAlertDefinition":{"actions":[],"flatCondition":[{"id":"c56d2530-d445-4436-ba94-9bce4004c9a0"},{"id":"26641dd3-92a1-4479-b4bb-e46a5e5c8361"},{"id":"3e325c1d-f3fa-4054-aaa2-500c6156d9c6"},{"id":"ffe08f50-e05a-43c3-bc16-3c75486be7c9"},{"id":"f4167d89-d6e8-47f6-b084-82ee5378cb53"}],"description":"this is an alert to test nothing","enabled":false,"id":"43a1743c-91ca-43ee-a37e-df01902d2dc4","name":"mic-test-alert","organizationId":"140638900734749696","severity":"INFO"}}}}`
	readResponseJson   = `{"data":{"alertQueries":{"alertDefinitions":{"alertDefinitions":[{"actions":[],"triggerResetActions":false,"conditionType":"ENTITY_METRIC","flatCondition":[{"id":"935b93f6-f94f-4b25-98a6-e66bbf80eaee","links":[{"name":"operands","values":["0f9212ff-c437-4496-aabd-72a3d0c4dea0","9b3f1343-4936-40d2-bb2c-7ee8a7612f46"]}],"value":{"fieldName":null,"operator":">","type":"binaryOperator","query":null}},{"id":"0f9212ff-c437-4496-aabd-72a3d0c4dea0","links":[{"name":"operands","values":["8fc84a11-dece-4561-9b36-573c4e38929f","fa655d66-673c-444c-b368-4c173585699d"]}],"value":{"fieldName":null,"operator":"MAX","type":"aggregationOperator","query":null}},{"id":"9b3f1343-4936-40d2-bb2c-7ee8a7612f46","links":[],"value":{"fieldName":null,"operator":null,"type":"constantValue","query":null}},{"id":"8fc84a11-dece-4561-9b36-573c4e38929f","links":[],"value":{"fieldName":"Orion.NPM.InterfaceTraffic.InTotalBytes","operator":null,"type":"metricField","query":null}},{"id":"fa655d66-673c-444c-b368-4c173585699d","links":[],
	"value":{"fieldName":null,"operator":null,"type":"constantValue","query":null}}],"description":"this is an alert to test nothing","enabled":false,"id":"43a1743c-91ca-43ee-a37e-df01902d2dc4","name":"mic-test-alert","organizationId":"140638900734749696","severity":"INFO","triggered":false,"triggeredTime":null,"targetEntityTypes":["DeviceVolume"],"muteInfo":{"muted":false,"until":null},"userId":"151686710111094784"}]}}}}`
	updateResponseJson = `{"data":{"alertMutations":{"updateAlertDefinition":{"actions":[],"flatCondition":[{"id":"e6be7955-4dcd-4b4c-9cf1-f0161fec02fa"},{"id":"75e6e17b-2028-4c74-acc9-396d45456ccb"},{"id":"34c53802-e994-47de-adad-8bf394f36b1f"},{"id":"73c5560c-fef8-4a75-af0c-44c3ab7a89e7"},{"id":"78caaf1f-4be5-4d89-a061-cac4a535ee5c"}],"description":"this is an alert to test nothing","enabled":false,"id":"43a1743c-91ca-43ee-a37e-df01902d2dc4","muteInfo":{"muted":false,"until":null},"name":"mic-test-alert","organizationId":"140638900734749696","severity":"INFO","triggered":false,"triggeredTime":null,"targetEntityTypes":["DeviceVolume"],"userId":"151686710111094784"}}}}`
	deleteResponseJson = `{"data":{"alertMutations":{"deleteAlertDefinition":"43a1743c-91ca-43ee-a37e-df01902d2dc4"}}}`
	alertId            = "43a1743c-91ca-43ee-a37e-df01902d2dc4"
)

var (
	alertsService *AlertsService
)

func init() {
	client := Client{
		gql: graphql.NewClient("", &mocks.MockClient{}),
	}

	alertsService = NewAlertsService(&client)
}

func TestCreateAlert(t *testing.T) {
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		bodyData, err := io.ReadAll(req.Body)
		if err != nil {
			log.Fatalln(err)
		}

		readCloser := io.NopCloser(strings.NewReader(createResponseJson))
		//Checking the body length to confirm nothing changed when the gql request was created
		assert.Equal(t, 1725, len(bodyData))
		return &http.Response{
			StatusCode: 200,
			Body:       readCloser,
		}, nil
	}

	defaultAlertDefinition := DefaultAlertDefinition()
	resp, err := alertsService.Create(defaultAlertDefinition)
	if err != nil {
		log.Fatal(err)
	}

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, 451, len(jsonResp))
	assert.Equal(t, alertId, resp.Id)
	assert.Equal(t, defaultAlertDefinition.Name, resp.Name)
	assert.Equal(t, defaultAlertDefinition.Description, resp.Description)
	assert.Equal(t, defaultAlertDefinition.Enabled, resp.Enabled)

}

func TestReadAlert(t *testing.T) {
	addMockRequestBody(readResponseJson)

	resp, err := alertsService.Read(alertId)
	if err != nil {
		log.Fatal(err)
	}

	defaultAlertDefinition := DefaultAlertDefinition()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, 1370, len(jsonResp))
	assert.Equal(t, alertId, resp.Id)
	assert.Equal(t, defaultAlertDefinition.Name, resp.Name)
	assert.Equal(t, defaultAlertDefinition.Description, resp.Description)
	assert.Equal(t, defaultAlertDefinition.Enabled, resp.Enabled)
}

func TestUpdateAlert(t *testing.T) {
	addMockRequestBody(updateResponseJson)
	alertDefinition := DefaultAlertDefinition()

	ninetyEight := "98"
	alertDefinition.Condition[4].Value = &ninetyEight

	err := alertsService.Update(alertId, alertDefinition)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, err, nil)
}

func TestDeleteAlert(t *testing.T) {
	addMockRequestBody(deleteResponseJson)
	err := alertsService.Delete(alertId)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, err, nil)
}

func addMockRequestBody(responseJson string) {
	mocks.GetDoFunc = func(req *http.Request) (*http.Response, error) {
		readCloser := io.NopCloser(strings.NewReader(responseJson))
		return &http.Response{
			StatusCode: 200,
			Body:       readCloser,
		}, nil
	}
}

func DefaultAlertDefinition() AlertDefinitionInput {
	operatorGt := ">"
	operatorMax := "MAX"

	alertExpression := AlertFilterExpressionInput{Operation: FilterOperationEq}
	entityFilter := AlertConditionNodeEntityFilterInput{Type: "DeviceVolume"}

	dataTypeString := "string"
	dataTypeNumber := "number"

	oneDValue := "1d"
	Ninety := "90"

	fieldName := "Orion.NPM.InterfaceTraffic.InTotalBytes"

	alertCondition := []AlertConditionNodeInput{
		{Id: 1, Type: "binaryOperator", Operator: &operatorGt, OperandIds: []int{2, 5}},
		{Id: 2, Type: "aggregationOperator", Operator: &operatorMax, OperandIds: []int{3, 4}, MetricFilter: &alertExpression},
		{Id: 3, Type: "metricField", EntityFilter: &entityFilter, FieldName: &fieldName},
		{Id: 4, Type: "constantValue", DataType: &dataTypeString, Value: &oneDValue},
		{Id: 5, Type: "constantValue", DataType: &dataTypeNumber, Value: &Ninety},
	}

	description := "this is an alert to test nothing"

	return AlertDefinitionInput{
		Name:        "mic-test-alert",
		Description: &description,
		Enabled:     false,
		Condition:   alertCondition,
		Severity:    AlertSeverityInfo,
	}
}
