package newrelicinfra

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	newrelic "github.com/paul91/go-newrelic-infra/api"
)

// thresholdSchema returns the schema to use for threshold.
//
func thresholdSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"value": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"time_function": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"any", "all"}, false),
			},
		},
	}
}

func resourceNewRelicInfraAlertCondition() *schema.Resource {

	return &schema.Resource{
		Create: resourceNewRelicInfraAlertConditionCreate,
		Read:   resourceNewRelicInfraAlertConditionRead,
		Update: resourceNewRelicInfraAlertConditionUpdate,
		Delete: resourceNewRelicInfraAlertConditionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"policy_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"infra_process_running", "infra_metric", "infra_host_not_reporting"}, false),
			},
			"event": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"where": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"comparison": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"above", "below", "equal"}, false),
			},
			"select": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"created_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"critical": {
				Type:     schema.TypeList,
				MinItems: 1,
				MaxItems: 1,
				Optional: true,
				Elem:     thresholdSchema(),
			},
			"warning": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				ForceNew: true,
				Elem:     thresholdSchema(),
			},
		},
	}
}

func buildInfraAlertConditionStruct(d *schema.ResourceData) *newrelic.AlertInfraCondition {

	condition := newrelic.AlertInfraCondition{
		Name:       d.Get("name").(string),
		Enabled:    d.Get("enabled").(bool),
		PolicyID:   d.Get("policy_id").(int),
		Event:      d.Get("event").(string),
		Comparison: d.Get("comparison").(string),
		Select:     d.Get("select").(string),
		Type:       d.Get("type").(string),
		Critical:   expandAlertThreshold(d.Get("critical")),
	}

	if attr, ok := d.GetOk("warning"); ok {
		condition.Warning = expandAlertThreshold(attr)
	}

	if v, ok := d.GetOk("enabled"); ok {
		condition.Enabled = v.(bool)
	}

	if v, ok := d.GetOk("where"); ok {
		condition.Where = v.(string)
	}

	return &condition
}

func readInfraAlertConditionStruct(condition *newrelic.AlertInfraCondition, d *schema.ResourceData) error {
	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]

	d.Set("policy_id", policyID)
	d.Set("name", condition.Name)
	d.Set("enabled", condition.Enabled)
	d.Set("created_at", condition.CreatedAt)
	d.Set("updated_at", condition.UpdatedAt)

	if condition.Where != "" {
		d.Set("where", condition.Where)
	}

	if err := d.Set("critical", flattenAlertThreshold(condition.Critical)); err != nil {
		return err
	}

	if condition.Warning != nil {
		if err := d.Set("warning", flattenAlertThreshold(condition.Warning)); err != nil {
			return err
		}
	}

	return nil
}

func resourceNewRelicInfraAlertConditionCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)
	condition := buildInfraAlertConditionStruct(d)

	log.Printf("[INFO] Creating New Relic Infra alert condition %s", condition.Name)

	condition, err := client.CreateAlertInfraCondition(*condition)
	if err != nil {
		return err
	}

	d.SetId(serializeIDs([]int{condition.PolicyID, condition.ID}))

	return resourceNewRelicInfraAlertConditionRead(d, meta)
}

func resourceNewRelicInfraAlertConditionRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)

	log.Printf("[INFO] Reading New Relic Infra alert condition %s", d.Id())

	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]
	id := ids[1]

	condition, err := client.GetAlertInfraCondition(policyID, id)
	if err != nil {
		if err == newrelic.ErrNotFound {
			d.SetId("")
			return nil
		}

		return err
	}

	return readInfraAlertConditionStruct(condition, d)
}

func resourceNewRelicInfraAlertConditionUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)
	condition := buildInfraAlertConditionStruct(d)

	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]
	id := ids[1]

	condition.PolicyID = policyID
	condition.ID = id

	log.Printf("[INFO] Updating New Relic Infra alert condition %d", id)

	_, err = client.UpdateAlertInfraCondition(*condition)
	if err != nil {
		return err
	}

	return resourceNewRelicInfraAlertConditionRead(d, meta)
}

func resourceNewRelicInfraAlertConditionDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*newrelic.Client)

	ids, err := parseIDs(d.Id(), 2)
	if err != nil {
		return err
	}

	policyID := ids[0]
	id := ids[1]

	log.Printf("[INFO] Deleting New Relic Infra alert condition %d", id)

	if err := client.DeleteAlertInfraCondition(policyID, id); err != nil {
		return err
	}

	d.SetId("")

	return nil
}

func expandAlertThreshold(v interface{}) *newrelic.AlertInfraThreshold {
	rah := v.([]interface{})[0].(map[string]interface{})
	alertInfraThreshold := &newrelic.AlertInfraThreshold{
		Duration: rah["duration"].(int),
	}

	if val, ok := rah["value"]; ok {
		alertInfraThreshold.Value = val.(int)
	}

	if val, ok := rah["time_function"]; ok {
		alertInfraThreshold.Function = val.(string)
	}

	return alertInfraThreshold
}

func flattenAlertThreshold(v *newrelic.AlertInfraThreshold) []interface{} {
	alertInfraThreshold := map[string]interface{}{
		"duration":      v.Duration,
		"value":         v.Value,
		"time_function": v.Function,
	}

	return []interface{}{alertInfraThreshold}
}
