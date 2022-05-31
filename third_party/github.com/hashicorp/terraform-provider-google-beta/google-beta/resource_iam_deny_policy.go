// ----------------------------------------------------------------------------
//
//     ***     AUTO GENERATED CODE    ***    Type: MMv1     ***
//
// ----------------------------------------------------------------------------
//
//     This file is automatically generated by Magic Modules and manual
//     changes will be clobbered when the file is regenerated.
//
//     Please read more about how to change this file in
//     .github/CONTRIBUTING.md.
//
// ----------------------------------------------------------------------------

package google

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIAM2DenyPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceIAM2DenyPolicyCreate,
		Read:   resourceIAM2DenyPolicyRead,
		Update: resourceIAM2DenyPolicyUpdate,
		Delete: resourceIAM2DenyPolicyDelete,

		Importer: &schema.ResourceImporter{
			State: resourceIAM2DenyPolicyImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `The name of the policy.`,
			},
			"parent": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `The attachment point is identified by its URL-encoded full resource name.`,
			},
			"rules": {
				Type:        schema.TypeList,
				Required:    true,
				Description: `Rules to be applied.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"deny_rule": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: `A deny rule in an IAM deny policy.`,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"denial_condition": {
										Type:        schema.TypeList,
										Required:    true,
										Description: `User defined CEVAL expression. A CEVAL expression is used to specify match criteria such as origin.ip, source.region_code and contents in the request header.`,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"expression": {
													Type:        schema.TypeString,
													Required:    true,
													Description: `Textual representation of an expression in Common Expression Language syntax.`,
												},
												"description": {
													Type:     schema.TypeString,
													Optional: true,
													Description: `Description of the expression. This is a longer text which describes the expression,
e.g. when hovered over it in a UI.`,
												},
												"location": {
													Type:     schema.TypeString,
													Optional: true,
													Description: `String indicating the location of the expression for error reporting,
e.g. a file name and a position in the file.`,
												},
												"title": {
													Type:     schema.TypeString,
													Optional: true,
													Description: `Title for the expression, i.e. a short string describing its purpose.
This can be used e.g. in UIs which allow to enter the expression.`,
												},
											},
										},
									},
									"denied_permissions": {
										Type:     schema.TypeList,
										Optional: true,
										Description: `The permissions that are explicitly denied by this rule. Each permission uses the format '{service-fqdn}/{resource}.{verb}',
where '{service-fqdn}' is the fully qualified domain name for the service. For example, 'iam.googleapis.com/roles.list'.`,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"denied_principals": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: `The identities that are prevented from using one or more permissions on Google Cloud resources.`,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"exception_permissions": {
										Type:     schema.TypeList,
										Optional: true,
										Description: `Specifies the permissions that this rule excludes from the set of denied permissions given by deniedPermissions.
If a permission appears in deniedPermissions and in exceptionPermissions then it will not be denied.
The excluded permissions can be specified using the same syntax as deniedPermissions.`,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"exception_principals": {
										Type:     schema.TypeList,
										Optional: true,
										Description: `The identities that are excluded from the deny rule, even if they are listed in the deniedPrincipals.
For example, you could add a Google group to the deniedPrincipals, then exclude specific users who belong to that group.`,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: `The description of the rule.`,
						},
					},
				},
			},
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: `The display name of the rule.`,
			},
			"etag": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: `The hash of the resource. Used internally during updates.`,
			},
		},
		UseJSONNumber: true,
	}
}

func resourceIAM2DenyPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return err
	}

	obj := make(map[string]interface{})
	displayNameProp, err := expandIAM2DenyPolicyDisplayName(d.Get("display_name"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("display_name"); !isEmptyValue(reflect.ValueOf(displayNameProp)) && (ok || !reflect.DeepEqual(v, displayNameProp)) {
		obj["displayName"] = displayNameProp
	}
	etagProp, err := expandIAM2DenyPolicyEtag(d.Get("etag"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("etag"); !isEmptyValue(reflect.ValueOf(etagProp)) && (ok || !reflect.DeepEqual(v, etagProp)) {
		obj["etag"] = etagProp
	}
	rulesProp, err := expandIAM2DenyPolicyRules(d.Get("rules"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("rules"); !isEmptyValue(reflect.ValueOf(rulesProp)) && (ok || !reflect.DeepEqual(v, rulesProp)) {
		obj["rules"] = rulesProp
	}

	url, err := replaceVars(d, config, "{{IAM2BasePath}}policies/{{parent}}/denypolicies?policyId={{name}}")
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Creating new DenyPolicy: %#v", obj)
	billingProject := ""

	// err == nil indicates that the billing_project value was found
	if bp, err := getBillingProject(d, config); err == nil {
		billingProject = bp
	}

	res, err := sendRequestWithTimeout(config, "POST", billingProject, url, userAgent, obj, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("Error creating DenyPolicy: %s", err)
	}

	// Store the ID now
	id, err := replaceVars(d, config, "{{parent}}/{{name}}")
	if err != nil {
		return fmt.Errorf("Error constructing id: %s", err)
	}
	d.SetId(id)

	err = iAM2OperationWaitTime(
		config, res, "Creating DenyPolicy", userAgent,
		d.Timeout(schema.TimeoutCreate))

	if err != nil {
		// The resource didn't actually create
		d.SetId("")
		return fmt.Errorf("Error waiting to create DenyPolicy: %s", err)
	}

	log.Printf("[DEBUG] Finished creating DenyPolicy %q: %#v", d.Id(), res)

	return resourceIAM2DenyPolicyRead(d, meta)
}

func resourceIAM2DenyPolicyRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return err
	}

	url, err := replaceVars(d, config, "{{IAM2BasePath}}policies/{{parent}}/denypolicies/{{name}}")
	if err != nil {
		return err
	}

	billingProject := ""

	// err == nil indicates that the billing_project value was found
	if bp, err := getBillingProject(d, config); err == nil {
		billingProject = bp
	}

	res, err := sendRequest(config, "GET", billingProject, url, userAgent, nil)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("IAM2DenyPolicy %q", d.Id()))
	}

	if err := d.Set("display_name", flattenIAM2DenyPolicyDisplayName(res["displayName"], d, config)); err != nil {
		return fmt.Errorf("Error reading DenyPolicy: %s", err)
	}
	if err := d.Set("etag", flattenIAM2DenyPolicyEtag(res["etag"], d, config)); err != nil {
		return fmt.Errorf("Error reading DenyPolicy: %s", err)
	}
	if err := d.Set("rules", flattenIAM2DenyPolicyRules(res["rules"], d, config)); err != nil {
		return fmt.Errorf("Error reading DenyPolicy: %s", err)
	}

	return nil
}

func resourceIAM2DenyPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return err
	}

	billingProject := ""

	obj := make(map[string]interface{})
	displayNameProp, err := expandIAM2DenyPolicyDisplayName(d.Get("display_name"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("display_name"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, displayNameProp)) {
		obj["displayName"] = displayNameProp
	}
	etagProp, err := expandIAM2DenyPolicyEtag(d.Get("etag"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("etag"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, etagProp)) {
		obj["etag"] = etagProp
	}
	rulesProp, err := expandIAM2DenyPolicyRules(d.Get("rules"), d, config)
	if err != nil {
		return err
	} else if v, ok := d.GetOkExists("rules"); !isEmptyValue(reflect.ValueOf(v)) && (ok || !reflect.DeepEqual(v, rulesProp)) {
		obj["rules"] = rulesProp
	}

	url, err := replaceVars(d, config, "{{IAM2BasePath}}policies/{{parent}}/denypolicies/{{name}}")
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Updating DenyPolicy %q: %#v", d.Id(), obj)

	// err == nil indicates that the billing_project value was found
	if bp, err := getBillingProject(d, config); err == nil {
		billingProject = bp
	}

	res, err := sendRequestWithTimeout(config, "PUT", billingProject, url, userAgent, obj, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return fmt.Errorf("Error updating DenyPolicy %q: %s", d.Id(), err)
	} else {
		log.Printf("[DEBUG] Finished updating DenyPolicy %q: %#v", d.Id(), res)
	}

	err = iAM2OperationWaitTime(
		config, res, "Updating DenyPolicy", userAgent,
		d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return err
	}

	return resourceIAM2DenyPolicyRead(d, meta)
}

func resourceIAM2DenyPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	userAgent, err := generateUserAgentString(d, config.userAgent)
	if err != nil {
		return err
	}

	billingProject := ""

	url, err := replaceVars(d, config, "{{IAM2BasePath}}policies/{{parent}}/denypolicies/{{name}}")
	if err != nil {
		return err
	}

	var obj map[string]interface{}
	log.Printf("[DEBUG] Deleting DenyPolicy %q", d.Id())

	// err == nil indicates that the billing_project value was found
	if bp, err := getBillingProject(d, config); err == nil {
		billingProject = bp
	}

	res, err := sendRequestWithTimeout(config, "DELETE", billingProject, url, userAgent, obj, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return handleNotFoundError(err, d, "DenyPolicy")
	}

	err = iAM2OperationWaitTime(
		config, res, "Deleting DenyPolicy", userAgent,
		d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting DenyPolicy %q: %#v", d.Id(), res)
	return nil
}

func resourceIAM2DenyPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if err := parseImportId([]string{
		"(?P<parent>[^/]+)/(?P<name>[^/]+)",
	}, d, config); err != nil {
		return nil, err
	}

	// Replace import id for the resource id
	id, err := replaceVars(d, config, "{{parent}}/{{name}}")
	if err != nil {
		return nil, fmt.Errorf("Error constructing id: %s", err)
	}
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}

func flattenIAM2DenyPolicyDisplayName(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyEtag(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRules(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	if v == nil {
		return v
	}
	l := v.([]interface{})
	transformed := make([]interface{}, 0, len(l))
	for _, raw := range l {
		original := raw.(map[string]interface{})
		if len(original) < 1 {
			// Do not include empty json objects coming back from the api
			continue
		}
		transformed = append(transformed, map[string]interface{}{
			"description": flattenIAM2DenyPolicyRulesDescription(original["description"], d, config),
			"deny_rule":   flattenIAM2DenyPolicyRulesDenyRule(original["denyRule"], d, config),
		})
	}
	return transformed
}
func flattenIAM2DenyPolicyRulesDescription(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRule(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	if v == nil {
		return nil
	}
	original := v.(map[string]interface{})
	if len(original) == 0 {
		return nil
	}
	transformed := make(map[string]interface{})
	transformed["denied_principals"] =
		flattenIAM2DenyPolicyRulesDenyRuleDeniedPrincipals(original["deniedPrincipals"], d, config)
	transformed["exception_principals"] =
		flattenIAM2DenyPolicyRulesDenyRuleExceptionPrincipals(original["exceptionPrincipals"], d, config)
	transformed["denied_permissions"] =
		flattenIAM2DenyPolicyRulesDenyRuleDeniedPermissions(original["deniedPermissions"], d, config)
	transformed["exception_permissions"] =
		flattenIAM2DenyPolicyRulesDenyRuleExceptionPermissions(original["exceptionPermissions"], d, config)
	transformed["denial_condition"] =
		flattenIAM2DenyPolicyRulesDenyRuleDenialCondition(original["denialCondition"], d, config)
	return []interface{}{transformed}
}
func flattenIAM2DenyPolicyRulesDenyRuleDeniedPrincipals(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleExceptionPrincipals(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleDeniedPermissions(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleExceptionPermissions(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleDenialCondition(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	if v == nil {
		return nil
	}
	original := v.(map[string]interface{})
	if len(original) == 0 {
		return nil
	}
	transformed := make(map[string]interface{})
	transformed["expression"] =
		flattenIAM2DenyPolicyRulesDenyRuleDenialConditionExpression(original["expression"], d, config)
	transformed["title"] =
		flattenIAM2DenyPolicyRulesDenyRuleDenialConditionTitle(original["title"], d, config)
	transformed["description"] =
		flattenIAM2DenyPolicyRulesDenyRuleDenialConditionDescription(original["description"], d, config)
	transformed["location"] =
		flattenIAM2DenyPolicyRulesDenyRuleDenialConditionLocation(original["location"], d, config)
	return []interface{}{transformed}
}
func flattenIAM2DenyPolicyRulesDenyRuleDenialConditionExpression(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleDenialConditionTitle(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleDenialConditionDescription(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func flattenIAM2DenyPolicyRulesDenyRuleDenialConditionLocation(v interface{}, d *schema.ResourceData, config *Config) interface{} {
	return v
}

func expandIAM2DenyPolicyDisplayName(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyEtag(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRules(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	l := v.([]interface{})
	req := make([]interface{}, 0, len(l))
	for _, raw := range l {
		if raw == nil {
			continue
		}
		original := raw.(map[string]interface{})
		transformed := make(map[string]interface{})

		transformedDescription, err := expandIAM2DenyPolicyRulesDescription(original["description"], d, config)
		if err != nil {
			return nil, err
		} else if val := reflect.ValueOf(transformedDescription); val.IsValid() && !isEmptyValue(val) {
			transformed["description"] = transformedDescription
		}

		transformedDenyRule, err := expandIAM2DenyPolicyRulesDenyRule(original["deny_rule"], d, config)
		if err != nil {
			return nil, err
		} else if val := reflect.ValueOf(transformedDenyRule); val.IsValid() && !isEmptyValue(val) {
			transformed["denyRule"] = transformedDenyRule
		}

		req = append(req, transformed)
	}
	return req, nil
}

func expandIAM2DenyPolicyRulesDescription(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRule(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	l := v.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}
	raw := l[0]
	original := raw.(map[string]interface{})
	transformed := make(map[string]interface{})

	transformedDeniedPrincipals, err := expandIAM2DenyPolicyRulesDenyRuleDeniedPrincipals(original["denied_principals"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedDeniedPrincipals); val.IsValid() && !isEmptyValue(val) {
		transformed["deniedPrincipals"] = transformedDeniedPrincipals
	}

	transformedExceptionPrincipals, err := expandIAM2DenyPolicyRulesDenyRuleExceptionPrincipals(original["exception_principals"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedExceptionPrincipals); val.IsValid() && !isEmptyValue(val) {
		transformed["exceptionPrincipals"] = transformedExceptionPrincipals
	}

	transformedDeniedPermissions, err := expandIAM2DenyPolicyRulesDenyRuleDeniedPermissions(original["denied_permissions"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedDeniedPermissions); val.IsValid() && !isEmptyValue(val) {
		transformed["deniedPermissions"] = transformedDeniedPermissions
	}

	transformedExceptionPermissions, err := expandIAM2DenyPolicyRulesDenyRuleExceptionPermissions(original["exception_permissions"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedExceptionPermissions); val.IsValid() && !isEmptyValue(val) {
		transformed["exceptionPermissions"] = transformedExceptionPermissions
	}

	transformedDenialCondition, err := expandIAM2DenyPolicyRulesDenyRuleDenialCondition(original["denial_condition"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedDenialCondition); val.IsValid() && !isEmptyValue(val) {
		transformed["denialCondition"] = transformedDenialCondition
	}

	return transformed, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDeniedPrincipals(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleExceptionPrincipals(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDeniedPermissions(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleExceptionPermissions(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDenialCondition(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	l := v.([]interface{})
	if len(l) == 0 || l[0] == nil {
		return nil, nil
	}
	raw := l[0]
	original := raw.(map[string]interface{})
	transformed := make(map[string]interface{})

	transformedExpression, err := expandIAM2DenyPolicyRulesDenyRuleDenialConditionExpression(original["expression"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedExpression); val.IsValid() && !isEmptyValue(val) {
		transformed["expression"] = transformedExpression
	}

	transformedTitle, err := expandIAM2DenyPolicyRulesDenyRuleDenialConditionTitle(original["title"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedTitle); val.IsValid() && !isEmptyValue(val) {
		transformed["title"] = transformedTitle
	}

	transformedDescription, err := expandIAM2DenyPolicyRulesDenyRuleDenialConditionDescription(original["description"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedDescription); val.IsValid() && !isEmptyValue(val) {
		transformed["description"] = transformedDescription
	}

	transformedLocation, err := expandIAM2DenyPolicyRulesDenyRuleDenialConditionLocation(original["location"], d, config)
	if err != nil {
		return nil, err
	} else if val := reflect.ValueOf(transformedLocation); val.IsValid() && !isEmptyValue(val) {
		transformed["location"] = transformedLocation
	}

	return transformed, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDenialConditionExpression(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDenialConditionTitle(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDenialConditionDescription(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}

func expandIAM2DenyPolicyRulesDenyRuleDenialConditionLocation(v interface{}, d TerraformResourceData, config *Config) (interface{}, error) {
	return v, nil
}