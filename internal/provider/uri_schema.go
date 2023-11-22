package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/solarwinds/terraform-provider-swo/internal/envvar"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UriResourceModel is the main resource structure
type UriResourceModel struct {
	Id              types.String               `tfsdk:"id"`
	Name            types.String               `tfsdk:"name"`
	Host            types.String               `tfsdk:"host"`
	Options         UriResourceOptions         `tfsdk:"options"`
	TcpOptions      *UriResourceTcpOptions     `tfsdk:"tcp_options"`
	TestDefinitions UriResourceTestDefinitions `tfsdk:"test_definitions"`
}

// UriResourceOptions represents the options field in the main resource
type UriResourceOptions struct {
	IsPingEnabled types.Bool `tfsdk:"is_ping_enabled"`
	IsTcpEnabled  types.Bool `tfsdk:"is_tcp_enabled"`
}

// UriResourceTcpOptions represents the tcp_options field in the main resource
type UriResourceTcpOptions struct {
	Port           types.Int64  `tfsdk:"port"`
	StringToExpect types.String `tfsdk:"string_to_expect"`
	StringToSend   types.String `tfsdk:"string_to_send"`
}

// UriResourceTestDefinitions represents the test_definitions field in the main resource
type UriResourceTestDefinitions struct {
	TestFromLocation      types.String                `tfsdk:"test_from_location"`
	LocationOptions       []UriResourceProbeLocation  `tfsdk:"location_options"`
	TestIntervalInSeconds types.Int64                 `tfsdk:"test_interval_in_seconds"`
	PlatformOptions       *UriResourcePlatformOptions `tfsdk:"platform_options"`
}

// UriResourceProbeLocation represents location_options field in test_definitions
type UriResourceProbeLocation struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

// UriResourcePlatformOptions represents platform_options field in test_definitions
type UriResourcePlatformOptions struct {
	TestFromAll types.Bool `tfsdk:"test_from_all"`
	Platforms   []string   `tfsdk:"platforms"`
}

func (r *UriResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "UriResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s Uris.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "This is a computed ID provided by the backend.",
				Computed:    true,
			},
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
						Description: "The location type to test from [REGION|COUNTRY|CITY].",
						Required:    true,
					},
					"location_options": schema.SetNestedAttribute{
						Description: "The Website availability monitoring location options.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Description: "The Website availability monitoring location option type.",
									Required:    true,
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
						Required:    true,
					},
					"platform_options": schema.SingleNestedAttribute{
						Description: "The platform options for this Uri check.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"test_from_all": schema.BoolAttribute{
								Description: "Whether or not to test from all platforms.",
								Required:    true,
							},
							"platforms": schema.ListAttribute{
								Description: "The platforms to test from [AWS|AZURE].",
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
