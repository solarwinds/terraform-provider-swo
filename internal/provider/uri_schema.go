package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"
)

// uriResourceModel is the main resource structure
type uriResourceModel struct {
	Id              types.String                `tfsdk:"id"`
	Name            types.String                `tfsdk:"name"`
	Host            types.String                `tfsdk:"host"`
	Options         *uriResourceOptions         `tfsdk:"options"`
	TcpOptions      *uriResourceTcpOptions      `tfsdk:"tcp_options"`
	TestDefinitions *uriResourceTestDefinitions `tfsdk:"test_definitions"`
}

// uriResourceOptions represents the options field in the main resource
type uriResourceOptions struct {
	IsPingEnabled types.Bool `tfsdk:"is_ping_enabled"`
	IsTcpEnabled  types.Bool `tfsdk:"is_tcp_enabled"`
}

// uriResourceTcpOptions represents the tcp_options field in the main resource
type uriResourceTcpOptions struct {
	Port           types.Int64  `tfsdk:"port"`
	StringToExpect types.String `tfsdk:"string_to_expect"`
	StringToSend   types.String `tfsdk:"string_to_send"`
}

// uriResourceTestDefinitions represents the test_definitions field in the main resource
type uriResourceTestDefinitions struct {
	TestFromLocation      types.String                `tfsdk:"test_from_location"`
	LocationOptions       []uriResourceProbeLocation  `tfsdk:"location_options"`
	TestIntervalInSeconds types.Int64                 `tfsdk:"test_interval_in_seconds"`
	PlatformOptions       *uriResourcePlatformOptions `tfsdk:"platform_options"`
}

// uriResourceProbeLocation represents location_options field in test_definitions
type uriResourceProbeLocation struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

// uriResourcePlatformOptions represents platform_options field in test_definitions
type uriResourcePlatformOptions struct {
	TestFromAll types.Bool `tfsdk:"test_from_all"`
	Platforms   []string   `tfsdk:"platforms"`
}

func (r *uriResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing Uri uptime checks.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "The name of this Uri check.",
				Required:    true,
			},
			"host": schema.StringAttribute{
				Description: "The IP address or host name to monitor.",
				Required:    true,
			},
			"options": schema.SingleNestedAttribute{
				Description: "The options for this Uri check.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"is_ping_enabled": schema.BoolAttribute{
						Description: "Whether or not to enable ping monitoring.",
						Required:    true,
					},
					"is_tcp_enabled": schema.BoolAttribute{
						Description: "Whether or not to enable tcp monitoring.",
						Required:    true,
					},
				},
			},
			"tcp_options": schema.SingleNestedAttribute{
				Description: "The tcp options for this Uri check.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"port": schema.Int64Attribute{
						Description: "The port to use for tcp monitoring.",
						Required:    true,
					},
					"string_to_expect": schema.StringAttribute{
						Description: "The string to expect in the response.",
						Optional:    true,
					},
					"string_to_send": schema.StringAttribute{
						Description: "The string to send in the request.",
						Optional:    true,
					},
				},
			},
			"test_definitions": schema.SingleNestedAttribute{
				Description: "The test definitions for this Uri check.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"test_from_location": schema.StringAttribute{
						Description: "The location type to test from.",
						Required:    true,
						Validators: []validator.String{
							validators.SingleOption(
								swoClient.ProbeLocationTypeRegion,
								swoClient.ProbeLocationTypeCountry,
								swoClient.ProbeLocationTypeCity),
						},
					},
					"location_options": schema.SetNestedAttribute{
						Description: "The Website availability monitoring location options.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "The Website availability monitoring location option type.",
									Required:    true,
									Validators: []validator.String{
										validators.SingleOption(
											swoClient.ProbeLocationTypeRegion,
											swoClient.ProbeLocationTypeCountry,
											swoClient.ProbeLocationTypeCity),
									},
								},
								"value": schema.StringAttribute{
									Description: "The Website availability monitoring location option value.",
									Required:    true,
								},
							},
						},
					},
					"test_interval_in_seconds": schema.Int64Attribute{
						Description: "The interval to test in seconds.",
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(900),
					},
					"platform_options": schema.SingleNestedAttribute{
						Description: "The platform options for this Uri check.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"test_from_all": schema.BoolAttribute{
								Description: "Whether or not to test from all platforms.",
								Required:    true,
							},
							"platforms": schema.SetAttribute{
								Description: "The platforms to test from. Valid values are [`AWS`, `AZURE`].",
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
		},
	}
}
