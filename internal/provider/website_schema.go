package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/solarwindscloud/terraform-provider-swo/internal/envvar"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The main Website Resource model that is derived from the schema.
type WebsiteResourceModel struct {
	Id         types.String       `tfsdk:"id"`
	Name       types.String       `tfsdk:"name"`
	Url        types.String       `tfsdk:"url"`
	Monitoring *WebsiteMonitoring `tfsdk:"monitoring"`
}

type ProbeLocation struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

type PlatformOptions struct {
	TestFromAll types.Bool `tfsdk:"test_from_all"`
	Platforms   []string   `tfsdk:"platforms"`
}

type SslMonitoring struct {
	DaysPriorToExpiration          types.Int64 `tfsdk:"days_prior_to_expiration"`
	Enabled                        types.Bool  `tfsdk:"enabled"`
	IgnoreIntermediateCertificates types.Bool  `tfsdk:"ignore_intermediate_certificates"`
}

type WebsiteMonitoring struct {
	Options       MonitoringOptions      `tfsdk:"options"`
	Availability  AvailabilityMonitoring `tfsdk:"availability"`
	Rum           RumMonitoring          `tfsdk:"rum"`
	CustomHeaders []CustomHeader         `tfsdk:"custom_headers"`
}

type MonitoringOptions struct {
	IsAvailabilityActive types.Bool `tfsdk:"is_availability_active"`
	IsRumActive          types.Bool `tfsdk:"is_rum_active"`
}

type AvailabilityMonitoring struct {
	CheckForString        CheckForStringType `tfsdk:"check_for_string"`
	SSL                   SslMonitoring      `tfsdk:"ssl"`
	Protocols             []string           `tfsdk:"protocols"`
	TestFromLocation      types.String       `tfsdk:"test_from_location"`
	TestIntervalInSeconds types.Int64        `tfsdk:"test_interval_in_seconds"`
	LocationOptions       []ProbeLocation    `tfsdk:"location_options"`
	PlatformOptions       PlatformOptions    `tfsdk:"platform_options"`
}

type RumMonitoring struct {
	ApdexTimeInSeconds types.Int64  `tfsdk:"apdex_time_in_seconds"`
	Snippet            types.String `tfsdk:"snippet"`
	Spa                types.Bool   `tfsdk:"spa"`
}

type CustomHeader struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type AvailabilityTestValidations struct {
	SSLCertificates SslCertificates    `tfsdk:"ssl_certificates"`
	CheckForString  CheckForStringType `tfsdk:"check_for_string"`
}

type SslCertificates struct {
	Name        types.String `tfsdk:"name"`
	ValidTo     types.String `tfsdk:"valid_to"`
	Certificate types.String `tfsdk:"certificate"`
}

type CheckForStringType struct {
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
}

func (r *WebsiteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Trace(ctx, "WebsiteResource: Schema")

	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("A terraform resource for managing %s Websites.", envvar.AppName),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Website Id. This is a computed value provided by the backend.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Website name.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The Url to monitor.",
				Required:    true,
			},
			"monitoring": schema.SingleNestedAttribute{
				Description: "The Website monitoring settings.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"options": schema.SingleNestedAttribute{
						Description: "The Website monitoring options.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"is_availability_active": schema.BoolAttribute{
								Description: "Is availability monitoring active?",
								Required:    true,
							},
							"is_rum_active": schema.BoolAttribute{
								Description: "Is RUM monitoring active?",
								Required:    true,
							},
						},
					},
					"availability": schema.SingleNestedAttribute{
						Description: "The Website availability monitoring settings.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"check_for_string": schema.SingleNestedAttribute{
								Description: "The Website availability monitoring check for string settings.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"operator": schema.StringAttribute{
										Description: "The Website availability monitoring check for string operator.",
										Required:    true,
									},
									"value": schema.StringAttribute{
										Description: "The Website availability monitoring check for string value.",
										Required:    true,
									},
								},
							},
							"ssl": schema.SingleNestedAttribute{
								Description: "The Website availability monitoring SSL settings.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"days_prior_to_expiration": schema.Int64Attribute{
										Description: "The Website availability monitoring SSL days prior to expiration.",
										Required:    true,
									},
									"enabled": schema.BoolAttribute{
										Description: "Is SSL monitoring enabled?",
										Required:    true,
									},
									"ignore_intermediate_certificates": schema.BoolAttribute{
										Description: "Ignore intermediate certificates?",
										Required:    true,
									},
								},
							},
							"protocols": schema.ListAttribute{
								Description: "The Website availability monitoring protocols.",
								Required:    true,
								ElementType: types.StringType,
							},
							"test_from_location": schema.StringAttribute{
								Description: "The Website availability monitoring test from location.",
								Required:    true,
							},
							"test_interval_in_seconds": schema.Int64Attribute{
								Description: "The Website availability monitoring test interval in seconds.",
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
							"platform_options": schema.SingleNestedAttribute{
								Description: "The Website availability monitoring platform options.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"test_from_all": schema.BoolAttribute{
										Description: "Test from all platforms?",
										Required:    true,
									},
									"platforms": schema.ListAttribute{
										Description: "The Website availability monitoring platform options.",
										Required:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
					"rum": schema.SingleNestedAttribute{
						Description: "The Website RUM monitoring settings.",
						Required:    true,
						Attributes: map[string]schema.Attribute{
							"apdex_time_in_seconds": schema.Int64Attribute{
								Description: "The Website RUM monitoring apdex time in seconds.",
								Required:    true,
							},
							"snippet": schema.StringAttribute{
								Description: "The Website RUM monitoring code snippet.",
								Optional:    true,
							},
							"spa": schema.BoolAttribute{
								Description: "Is SPA monitoring enabled?",
								Required:    true,
							},
						},
					},
					"custom_headers": schema.SetNestedAttribute{
						Description: "One or more custom headers to send with the uptime check.",
						Required:    true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Description: "The Website custom header name.",
									Required:    true,
								},
								"value": schema.StringAttribute{
									Description: "The Website custom header value.",
									Required:    true,
								},
							},
						},
					},
				},
			},
		},
	}
}
