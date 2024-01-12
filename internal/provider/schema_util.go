package provider

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func resourceIdAttribute() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "The Id of the resource provided by the backend.",
		Computed:    true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	}
}

func enrichFrameworkResourceSchema(s *schema.Schema) {
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichDescription(attr)
	}

	for _, block := range s.Blocks {
		switch v := block.(type) {
		case schema.ListNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDescription(attr)
			}
		case schema.SingleNestedBlock:
			for i, attr := range v.Attributes {
				v.Attributes[i] = enrichDescription(attr)
			}
		case schema.SetNestedBlock:
			for i, attr := range v.NestedObject.Attributes {
				v.NestedObject.Attributes[i] = enrichDescription(attr)
			}
		}
	}
}

func enrichDescription(r any) schema.Attribute {
	switch v := r.(type) {
	case schema.StringAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case schema.Int64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case schema.Float64Attribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	case schema.BoolAttribute:
		buildEnrichedSchemaDescription(reflect.ValueOf(&v))
		return v
	default:
		return r.(schema.Attribute)
	}
}

func buildEnrichedSchemaDescription(rv reflect.Value) {
	descField := rv.Elem().FieldByName("Description")
	curentDesc := descField.String()

	// Build description with validators
	validators := rv.Elem().FieldByName("Validators")
	if validators.IsValid() && !validators.IsNil() && validators.Len() > 0 {
		for i := 0; i < validators.Len(); i++ {
			validator := validators.Index(i)
			name := validator.Elem().Type().Name()

			if strings.HasPrefix(name, "singleOptionValidator") {
				values := validator.Elem().FieldByName("ValidValues")
				v := reflect.ValueOf(values.Interface())
				validValuesMsg := ""
				sep := ""
				for i := 0; i < v.Len(); i++ {
					if len(validValuesMsg) > 0 {
						sep = "|"
					}
					validValuesMsg += fmt.Sprintf("%s%v", sep, v.Index(i))
				}
				curentDesc = fmt.Sprintf("%s Valid values are [%s].", curentDesc, validValuesMsg)
				break
			}
		}
	}

	// Build description with Defaults
	defaultField := rv.Elem().FieldByName("Default")
	if defaultField.IsValid() && !defaultField.IsNil() {
		defaultVal := defaultField.Elem().FieldByName("defaultVal")
		if defaultVal.IsValid() {
			curentDesc = fmt.Sprintf("%s Default is %v.", curentDesc, defaultVal)
		}
	}

	descField.SetString(curentDesc)
}
