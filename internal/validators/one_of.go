package validators

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/solarwinds/terraform-provider-swo/internal/typex"
)

func OneOf[T any](values ...T) validator.String {
	return stringvalidator.OneOf(typex.Map(values, func(v T) string {
		return fmt.Sprint(v)
	})...)
}
