package client

import (
	"log"
)

type AlertDefinition struct {
	ID                  string                          `json:"id"`
	Name                string                          `json:"name"`
	Description         string                          `json:"description"`
	Severity            AlertSeverity                   `json:"severity"`
	Enabled             bool                            `json:"enabled"`
	FlatCondition       []*FlatAlertConditionExpression `json:"flatCondition"`
	Actions             []*AlertAction                  `json:"actions"`
	TriggerResetActions bool                            `json:"triggerResetActions"`
	OrganizationID      string                          `json:"organizationId"`
	Triggered           bool                            `json:"triggered"`
	TriggeredTime       string                          `json:"triggeredTime"`
	UserID              string                          `json:"userId"`
}

type AlertSeverity string

const (
	AlertSeverityCritical AlertSeverity = "CRITICAL"
	AlertSeverityInfo     AlertSeverity = "INFO"
	AlertSeverityWarning  AlertSeverity = "WARNING"
)

var AllAlertSeverity = []AlertSeverity{
	AlertSeverityCritical,
	AlertSeverityInfo,
	AlertSeverityWarning,
}

func (e AlertSeverity) IsValid() bool {
	switch e {
	case AlertSeverityCritical, AlertSeverityInfo, AlertSeverityWarning:
		return true
	}
	return false
}

type FlatAlertConditionExpression struct {
	ID    string                  `json:"id"`
	Links []*NamedLinks           `json:"links"`
	Value *FlatAlertConditionNode `json:"value"`
}

type NamedLinks struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type FlatAlertConditionNode struct {
	DataType     *string                         `json:"dataType"`
	EntityFilter *AlertConditionNodeEntityFilter `json:"entityFilter"`
	FieldName    *string                         `json:"fieldName"`
	MetricFilter []*FlatAlertFilterExpression    `json:"metricFilter"`
	Operator     *string                         `json:"operator"`
	Query        *string                         `json:"query"`
	Source       *string                         `json:"source"`
	Type         string                          `json:"type"`
	Value        *string                         `json:"value"`
}

type AlertConditionNodeEntityFilter struct {
	Fields []*AlertConditionMatchFieldRule `json:"fields"`
	Ids    []string                        `json:"ids"`
	Type   string                          `json:"type"`
	Types  []string                        `json:"types"`
}

type AlertConditionMatchFieldRule struct {
	FieldName string                     `json:"fieldName"`
	Rules     []*AlertConditionMatchRule `json:"rules"`
}

type AlertConditionMatchRule struct {
	Negate bool                        `json:"negate"`
	Type   AlertConditionMatchRuleType `json:"type"`
	Value  string                      `json:"value"`
}

type AlertConditionMatchRuleType string

const (
	AlertConditionMatchRuleTypeContains AlertConditionMatchRuleType = "CONTAINS"
	AlertConditionMatchRuleTypeEq       AlertConditionMatchRuleType = "EQ"
	AlertConditionMatchRuleTypeMatches  AlertConditionMatchRuleType = "MATCHES"
	AlertConditionMatchRuleTypeNe       AlertConditionMatchRuleType = "NE"
)

var AllAlertConditionMatchRuleType = []AlertConditionMatchRuleType{
	AlertConditionMatchRuleTypeContains,
	AlertConditionMatchRuleTypeEq,
	AlertConditionMatchRuleTypeMatches,
	AlertConditionMatchRuleTypeNe,
}

func (e AlertConditionMatchRuleType) IsValid() bool {
	switch e {
	case AlertConditionMatchRuleTypeContains, AlertConditionMatchRuleTypeEq, AlertConditionMatchRuleTypeMatches, AlertConditionMatchRuleTypeNe:
		return true
	}
	return false
}

type FlatAlertFilterExpression struct {
	ID    string                 `json:"id"`
	Links []*NamedLinks          `json:"links"`
	Value *AlertFilterExpression `json:"value"`
}

type FlatEvaluatedConditionTreeNode struct {
	ID    string             `json:"id"`
	Links []*NamedLinks      `json:"links"`
	Value *EvaluatedTreeNode `json:"value"`
}

type AlertFilterExpression struct {
	Operation      FilterOperation `json:"operation"`
	PropertyName   *string         `json:"propertyName"`
	PropertyValue  *string         `json:"propertyValue"`
	PropertyValues []*string       `json:"propertyValues"`
}

type FilterOperation string

const (
	FilterOperationAnd      FilterOperation = "AND"
	FilterOperationContains FilterOperation = "CONTAINS"
	FilterOperationEq       FilterOperation = "EQ"
	FilterOperationExists   FilterOperation = "EXISTS"
	FilterOperationGe       FilterOperation = "GE"
	FilterOperationGt       FilterOperation = "GT"
	FilterOperationIn       FilterOperation = "IN"
	FilterOperationLe       FilterOperation = "LE"
	FilterOperationLt       FilterOperation = "LT"
	FilterOperationMatches  FilterOperation = "MATCHES"
	FilterOperationNe       FilterOperation = "NE"
	FilterOperationNot      FilterOperation = "NOT"
	FilterOperationOr       FilterOperation = "OR"
)

var AllFilterOperation = []FilterOperation{
	FilterOperationAnd,
	FilterOperationContains,
	FilterOperationEq,
	FilterOperationExists,
	FilterOperationGe,
	FilterOperationGt,
	FilterOperationIn,
	FilterOperationLe,
	FilterOperationLt,
	FilterOperationMatches,
	FilterOperationNe,
	FilterOperationNot,
	FilterOperationOr,
}

func (e FilterOperation) IsValid() bool {
	switch e {
	case FilterOperationAnd, FilterOperationContains, FilterOperationEq, FilterOperationExists, FilterOperationGe, FilterOperationGt, FilterOperationIn, FilterOperationLe, FilterOperationLt, FilterOperationMatches, FilterOperationNe, FilterOperationNot, FilterOperationOr:
		return true
	}
	return false
}

type EvaluatedTreeNode struct {
	FieldName *string          `json:"fieldName"`
	Operator  *string          `json:"operator"`
	Query     *string          `json:"query"`
	Result    *EvaluatedResult `json:"result"`
	Source    *string          `json:"source"`
	Type      string           `json:"type"`
}

type EvaluatedResult struct {
	Boolean *bool           `json:"boolean"`
	Number  *float64        `json:"number"`
	String  *string         `json:"string"`
	Type    *EvalResultType `json:"type"`
}

type EvalResultType string

const (
	EvalResultTypeBoolean EvalResultType = "BOOLEAN"
	EvalResultTypeNull    EvalResultType = "NULL"
	EvalResultTypeNumber  EvalResultType = "NUMBER"
	EvalResultTypeString  EvalResultType = "STRING"
)

var AllEvalResultType = []EvalResultType{
	EvalResultTypeBoolean,
	EvalResultTypeNull,
	EvalResultTypeNumber,
	EvalResultTypeString,
}

func (e EvalResultType) IsValid() bool {
	switch e {
	case EvalResultTypeBoolean, EvalResultTypeNull, EvalResultTypeNumber, EvalResultTypeString:
		return true
	}
	return false
}

type AlertAction struct {
	ConfigurationIds []string `json:"configurationIds"`
	Type             string   `json:"type"`
}

type AlertsService struct {
	client *Client
}

type AlertsCommunicator interface {
	Get(string) (*AlertDefinition, error)
	Create(*AlertDefinition) (*AlertDefinition, error)
	Update(*AlertDefinition) error
	Delete(string) error
}

func NewAlertsService(c *Client) *AlertsService {
	return &AlertsService{c}
}

// Returns the alert identified by the given Id.
func (as *AlertsService) Get(id string) (*AlertDefinition, error) {
	log.Println("Read alert request.")

	return &AlertDefinition{
		ID:          id,
		Name:        "mock name",
		Description: "mock description",
	}, nil

	// alert := &AlertDefinition{}
	// path := fmt.Sprintf("alerts/%s", id)
	// req, err := as.client.NewRequest("GET", path, nil)

	// if err != nil {
	// 	return nil, err
	// }

	// _, err = as.client.Do(req, alert)
	// if err != nil {
	// 	return nil, err
	// }

	// return alert, nil
}

// Creates a new alert with the given definition.
func (as *AlertsService) Create(a *AlertDefinition) (*AlertDefinition, error) {
	log.Println("Create alert request.")

	a.ID = "new_alert_id"
	return a, nil

	// req, err := as.client.NewRequest("POST", "alerts", a)
	// if err != nil {
	// 	return nil, err
	// }

	// createdAlert := &AlertDefinition{}

	// _, err = as.client.Do(req, createdAlert)
	// if err != nil {
	// 	return nil, err
	// }

	// return createdAlert, nil
}

// Updates the alert with the given id.
func (as *AlertsService) Update(a *AlertDefinition) error {
	log.Println("Update alert request.")

	return nil

	// path := fmt.Sprintf("alerts/%s", a.ID)
	// req, err := as.client.NewRequest("PUT", path, a)
	// if err != nil {
	// 	return err
	// }
	// _, err = as.client.Do(req, nil)
	// if err != nil {
	// 	return err
	// }
	// return nil
}

// Deletes the alert with the given id.
func (as *AlertsService) Delete(id string) error {
	log.Println("Delete alert request.")

	return nil

	// path := fmt.Sprintf("alerts/%s", id)
	// req, err := as.client.NewRequest("DELETE", path, nil)
	// if err != nil {
	// 	return err
	// }

	// _, err = as.client.Do(req, nil)

	// return err
}
