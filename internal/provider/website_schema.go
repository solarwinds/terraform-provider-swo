package provider

import (
	"context"
	"github.com/solarwinds/swo-sdk-go/swov1/models/components"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	swoClient "github.com/solarwinds/swo-client-go/pkg/client"
	"github.com/solarwinds/terraform-provider-swo/internal/validators"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// The main Website Resource model that is derived from the schema.
type websiteResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Url        types.String `tfsdk:"url"`
	Tags       types.Set    `tfsdk:"tags"`
	Monitoring types.Object `tfsdk:"monitoring"` //websiteMonitoring
}

type websiteTag struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func WebsiteTagAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"key":   types.StringType,
		"value": types.StringType,
	}
}

type websiteMonitoring struct {
	Options       types.Object `tfsdk:"options"`        //monitoringOptions
	Availability  types.Object `tfsdk:"availability"`   //availabilityMonitoring
	Rum           types.Object `tfsdk:"rum"`            //rumMonitoring
	CustomHeaders types.Set    `tfsdk:"custom_headers"` //deprecated
}

func WebsiteMonitoringAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"options":        types.ObjectType{AttrTypes: MonitoringOptionsAttributeTypes()},
		"availability":   types.ObjectType{AttrTypes: AvailabilityMonitoringAttributeTypes()},
		"rum":            types.ObjectType{AttrTypes: RumMonitoringAttributeTypes()},
		"custom_headers": types.SetType{ElemType: types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}},
	}
}

type monitoringOptions struct {
	IsAvailabilityActive types.Bool `tfsdk:"is_availability_active"`
	IsRumActive          types.Bool `tfsdk:"is_rum_active"`
}

func MonitoringOptionsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"is_availability_active": types.BoolType,
		"is_rum_active":          types.BoolType,
	}
}

type availabilityMonitoring struct {
	CheckForString        types.Object `tfsdk:"check_for_string"`
	SSL                   types.Object `tfsdk:"ssl"`
	Protocols             types.List   `tfsdk:"protocols"`
	TestFromLocation      types.String `tfsdk:"test_from_location"`
	TestIntervalInSeconds types.Int64  `tfsdk:"test_interval_in_seconds"`
	LocationOptions       types.Set    `tfsdk:"location_options"`
	PlatformOptions       types.Object `tfsdk:"platform_options"`
	CustomHeaders         types.Set    `tfsdk:"custom_headers"`
	OutageConfig          types.Object `tfsdk:"outage_configuration"`
}

func AvailabilityMonitoringAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"check_for_string":         types.ObjectType{AttrTypes: CheckForStringTypeAttributeTypes()},
		"ssl":                      types.ObjectType{AttrTypes: SslMonitoringAttributeTypes()},
		"protocols":                types.ListType{ElemType: types.StringType},
		"test_from_location":       types.StringType,
		"test_interval_in_seconds": types.Int64Type,
		"location_options":         types.SetType{ElemType: types.ObjectType{AttrTypes: ProbeLocationAttributeTypes()}},
		"platform_options":         types.ObjectType{AttrTypes: PlatformOptionsAttributeTypes()},
		"custom_headers":           types.SetType{ElemType: types.ObjectType{AttrTypes: CustomHeaderAttributeTypes()}},
		"outage_configuration":     types.ObjectType{AttrTypes: OutageConfigAttributeTypes()},
	}
}

type rumMonitoring struct {
	ApdexTimeInSeconds types.Int64  `tfsdk:"apdex_time_in_seconds"`
	Snippet            types.String `tfsdk:"snippet"`
	Spa                types.Bool   `tfsdk:"spa"`
}

func RumMonitoringAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"apdex_time_in_seconds": types.Int64Type,
		"snippet":               types.StringType,
		"spa":                   types.BoolType,
	}
}

type probeLocation struct {
	Type  types.String `tfsdk:"type"`
	Value types.String `tfsdk:"value"`
}

func ProbeLocationAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"type":  types.StringType,
		"value": types.StringType,
	}
}

type platformOptions struct {
	TestFromAll types.Bool `tfsdk:"test_from_all"`
	Platforms   types.Set  `tfsdk:"platforms"`
}

func PlatformOptionsAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"test_from_all": types.BoolType,
		"platforms":     types.SetType{ElemType: types.StringType},
	}
}

type sslMonitoring struct {
	DaysPriorToExpiration          types.Int64 `tfsdk:"days_prior_to_expiration"`
	Enabled                        types.Bool  `tfsdk:"enabled"`
	IgnoreIntermediateCertificates types.Bool  `tfsdk:"ignore_intermediate_certificates"`
}

func SslMonitoringAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"days_prior_to_expiration":         types.Int64Type,
		"enabled":                          types.BoolType,
		"ignore_intermediate_certificates": types.BoolType,
	}
}

type customHeader struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func CustomHeaderAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":  types.StringType,
		"value": types.StringType,
	}
}

type outageConfig struct {
	FailingTestLocations types.String `tfsdk:"failing_test_locations"`
	ConsecutiveForDown   types.Int64  `tfsdk:"consecutive_for_down"`
}

func OutageConfigAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"failing_test_locations": types.StringType,
		"consecutive_for_down":   types.Int64Type,
	}
}

type checkForStringType struct {
	Operator types.String `tfsdk:"operator"`
	Value    types.String `tfsdk:"value"`
}

func CheckForStringTypeAttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"operator": types.StringType,
		"value":    types.StringType,
	}
}

func (r *websiteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform resource for managing website uptime checks.",
		Attributes: map[string]schema.Attribute{
			"id": resourceIdAttribute(),
			"name": schema.StringAttribute{
				Description: "Website name.",
				Required:    true,
			},
			"url": schema.StringAttribute{
				Description: "The Url to monitor.",
				Required:    true,
			},
			"tags": schema.SetNestedAttribute{
				Description: "Entity tags. Tag is a key-value pair, where there may be only single tag value for the same key.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							Required: true,
						},
						"value": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"monitoring": schema.SingleNestedAttribute{
				Description: "The Website monitoring settings.",
				Required:    true,
				Validators: []validator.Object{
					objectvalidator.AtLeastOneOf(path.Expressions{
						path.MatchRelative().AtName("availability"),
						path.MatchRelative().AtName("rum"),
					}...),
				},
				Attributes: map[string]schema.Attribute{
					"options": schema.SingleNestedAttribute{
						Description: "The Website monitoring options.",
						Computed:    true,
						Attributes: map[string]schema.Attribute{
							"is_availability_active": schema.BoolAttribute{
								Description: "Is availability monitoring active?",
								Computed:    true,
							},
							"is_rum_active": schema.BoolAttribute{
								Description: "Is RUM monitoring active?",
								Computed:    true,
							},
						},
					},
					"availability": schema.SingleNestedAttribute{
						Description: "The Website availability monitoring settings.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"check_for_string": schema.SingleNestedAttribute{
								Description: "The Website availability monitoring check for string settings.",
								Optional:    true,
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
								Optional:    true,
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
								Validators: []validator.String{
									validators.SingleOption(
										swoClient.ProbeLocationTypeRegion,
										swoClient.ProbeLocationTypeCountry,
										swoClient.ProbeLocationTypeCity),
								},
							},
							"test_interval_in_seconds": schema.Int64Attribute{
								Description: "The Website availability monitoring test interval in seconds. Valid values are 60, 300, 600, 900, 1800, 3600, 7200, 14400",
								Required:    true,
								Validators: []validator.Int64{
									int64validator.OneOf(60, 300, 600, 900, 1800, 3600, 7200, 14400),
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
							"platform_options": schema.SingleNestedAttribute{
								Description: "The Website availability monitoring platform options.",
								Required:    true,
								Attributes: map[string]schema.Attribute{
									"test_from_all": schema.BoolAttribute{
										Description: "Test from all platforms?",
										Required:    true,
									},
									"platforms": schema.SetAttribute{
										Description: "The Website availability monitoring platform options. Valid values are [AWS, AZURE, GOOGLE_CLOUD].",
										Required:    true,
										ElementType: types.StringType,
									},
								},
							},
							"custom_headers": schema.SetNestedAttribute{
								Description: "One or more custom headers to send with the uptime check.",
								Optional:    true,
								Validators: []validator.Set{
									setvalidator.ExactlyOneOf(path.Expressions{
										path.MatchRoot("monitoring").AtName("custom_headers"),
									}...),
								},
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
							"outage_configuration": schema.SingleNestedAttribute{
								Description: "Default conditions when the entity is considered down. " +
									"If omitted or set to null, organization configuration will be used for this entity.",
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"failing_test_locations": schema.StringAttribute{
										Description: "How many locations must report a failure for an entity to be considered down. Valid values are [all, any].",
										Required:    true,
										Validators: []validator.String{
											validators.SingleOption(
												components.WebsiteFailingTestLocationsAll,
												components.WebsiteFailingTestLocationsAny,
											),
										},
									},
									"consecutive_for_down": schema.Int64Attribute{
										Description: "Number of consecutive failing tests for an entity to be considered down. Minimum 1.",
										Required:    true,
									},
								},
							},
						},
					},
					"rum": schema.SingleNestedAttribute{
						Description: "The Website RUM monitoring settings.",
						Optional:    true,
						Attributes: map[string]schema.Attribute{
							"apdex_time_in_seconds": schema.Int64Attribute{
								Description: "The Website RUM monitoring apdex time in seconds.",
								Required:    true,
							},
							"snippet": schema.StringAttribute{
								Description: "The Website RUM monitoring code snippet (provided by the server).",
								Computed:    true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"spa": schema.BoolAttribute{
								Description: "Is SPA monitoring enabled?",
								Required:    true,
							},
						},
					},
					"custom_headers": schema.SetNestedAttribute{
						Description: "One or more custom headers to send with the uptime check. " +
							"custom_headers has been moved into monitoring.availability. " +
							"If this field and monitoring.availability.custom_headers are both set an error with be thrown. " +
							"If this field is set availability must also be set or an error will be thrown.",
						DeprecationMessage: "custom_headers has been moved into monitoring.availability. " +
							"Remove this attribute's configuration as it's no longer in use and the attribute will be removed in the next major version of the provider. " +
							"If this field and monitoring.availability.custom_headers are both set an error with be thrown. " +
							"If this field is set availability must also be set or an error will be thrown.",
						Optional: true,
						Validators: []validator.Set{
							setvalidator.AlsoRequires(path.Expressions{
								path.MatchRoot("monitoring").AtName("availability"),
							}...),
						},
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
