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
	for i, attr := range s.Attributes {
		s.Attributes[i] = enrichAttribute(attr)
	}
}

func enrichAttribute(attr schema.Attribute) schema.Attribute {
	switch v := attr.(type) {
	case schema.StringAttribute:
		return enrichDescription(v)
	case schema.BoolAttribute:
		return enrichDescription(v)
	case schema.Int64Attribute:
		return enrichDescription(v)
	case schema.Float64Attribute:
		return enrichDescription(v)
	case schema.NumberAttribute:
		return enrichDescription(v)
	case schema.ObjectAttribute:
		return enrichDescription(v)
	case schema.ListAttribute:
		return enrichDescription(v)
	case schema.MapAttribute:
		return enrichDescription(v)
	case schema.SetAttribute:
		return enrichDescription(v)
	case schema.SingleNestedAttribute:
		for i, chld := range v.Attributes {
			v.Attributes[i] = enrichAttribute(chld)
		}
		return enrichDescription(v)
	case schema.ListNestedAttribute:
		for i, chld := range v.NestedObject.Attributes {
			v.NestedObject.Attributes[i] = enrichAttribute(chld)
		}
		return enrichDescription(v)
	case schema.SetNestedAttribute:
		for i, chld := range v.NestedObject.Attributes {
			v.NestedObject.Attributes[i] = enrichAttribute(chld)
		}
		return enrichDescription(v)
	case schema.MapNestedAttribute:
		for i, chld := range v.NestedObject.Attributes {
			v.NestedObject.Attributes[i] = enrichAttribute(chld)
		}
		return enrichDescription(v)
	}

	return attr
}

func enrichDescription[T schema.Attribute](value T) T {
	rv := reflect.ValueOf(&value)
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
					validValuesMsg += fmt.Sprintf("%s`%v`", sep, v.Index(i))
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
			curentDesc = fmt.Sprintf("%s Default is `%v`.", curentDesc, defaultVal)
		}
	}

	descField.SetString(curentDesc)
	return value
}
