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

func enrichSchema(s *schema.Schema) {
	for _, attr := range s.Attributes {
		enrichResourceOrBlock(attr)
	}

	for _, block := range s.Blocks {
		enrichResourceOrBlock(block)
	}
}

// Enrich an attribute or block and all child attributes/blocks if the type has them.
// NOTE: This function is recursive.
func enrichResourceOrBlock(el any) {
	switch v := el.(type) {
	case schema.StringAttribute:
		enrichDescription(&v)
	case schema.BoolAttribute:
		enrichDescription(&v)
	case schema.Int64Attribute:
		enrichDescription(&v)
	case schema.ListAttribute:
		enrichDescription(&v)
	case schema.Float64Attribute:
		enrichDescription(&v)
	case schema.MapAttribute:
		enrichDescription(&v)
	case schema.SetAttribute:
		enrichDescription(&v)
	case schema.NumberAttribute:
		enrichDescription(&v)
	case schema.ObjectAttribute:
		enrichDescription(&v)
	case schema.SingleNestedAttribute:
		enrichNestedAttribute(v)
	case schema.ListNestedAttribute:
		enrichNestedAttribute(v)
		enrichDescription(&v)
	case schema.SetNestedAttribute:
		enrichNestedAttribute(v)
		enrichDescription(&v)
	case schema.MapNestedAttribute:
		enrichNestedAttribute(v)
		enrichDescription(&v)
	case schema.SingleNestedBlock:
		enrichBlock(v)
		enrichDescription(&v)
	case schema.ListNestedBlock:
		enrichBlock(v)
		enrichDescription(&v)
	case schema.SetNestedBlock:
		enrichBlock(v)
		enrichDescription(&v)
	}
}

func enrichNestedAttribute(attr schema.NestedAttribute) {
	for _, attr := range attr.GetNestedObject().GetAttributes() {
		enrichResourceOrBlock(attr)
	}
}

func enrichBlock(block schema.Block) {
	for _, attr := range block.GetNestedObject().GetAttributes() {
		enrichResourceOrBlock(attr)
	}
	for _, block := range block.GetNestedObject().GetBlocks() {
		enrichResourceOrBlock(block)
	}
}

func enrichDescription(value any) {
	rv := reflect.ValueOf(value)
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
				curentDesc = fmt.Sprintf("%s Valid values are `[%s]`.", curentDesc, validValuesMsg)
				break
			}
		}
	}

	// Build description with Defaults
	defaultField := rv.Elem().FieldByName("Default")
	if defaultField.IsValid() && !defaultField.IsNil() {
		defaultVal := defaultField.Elem().FieldByName("defaultVal")
		if defaultVal.IsValid() {
			curentDesc = fmt.Sprintf("%s Default is `%v`.", curentDesc, defaultVal)
		}
	}

	descField.SetString(curentDesc)
}
