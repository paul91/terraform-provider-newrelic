package newrelicinfra

import (
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	newrelic "github.com/paul91/go-newrelic-infra/api"
)

// thresholdSchema returns the schema to use for threshold.
//
func thresholdSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
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
				ValidateFunc: validation.StringInSlice([]string{"infra_process_running", "infra_metric", "infra_host_not_reporting"}, false),
			},
			"event": {
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
			"critical": thresholdSchema(),
		},
	}
}

func buildInfraAlertConditionStruct(d *schema.ResourceData) *newrelic.AlertInfraCondition {

	critical := newrelic.AlertInfraThreshold{}

	if valueQuery, ok := d.GetOk("critical.0.value"); ok {
		critical.Value = valueQuery.(int)
	}

	if durationValue, ok := d.GetOk("critical.0.duration"); ok {
		critical.Duration = durationValue.(int)
	}

	if functionValue, ok := d.GetOk("critical.0.time_function"); ok {
		critical.Function = functionValue.(string)
	}

	condition := newrelic.AlertInfraCondition{
		Name:       d.Get("name").(string),
		Enabled:    d.Get("enabled").(bool),
		PolicyID:   d.Get("policy_id").(int),
		Event:      d.Get("event").(string),
		Comparison: d.Get("comparison").(string),
		Select:     d.Get("select").(string),
		Type:       d.Get("type").(string),
		Critical:   critical,
	}

	if v, ok := d.GetOk("enabled"); ok {
		condition.Enabled = v.(bool)
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
