package newrelicinfra

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	newrelic "github.com/paul91/go-newrelic-infra/api"
)

func TestAccNewRelicInfraAlertCondition_Basic(t *testing.T) {
	rName := acctest.RandString(5)
	whereClause := "hostname LIKE 'frontend'"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNewRelicInfraAlertConditionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckNewRelicInfraAlertConditionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicInfraAlertConditionExists("newrelicinfra_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "name", fmt.Sprintf("tf-test-%s", rName)),
					// For some reason Enabled isn't being sent to the API?
					// resource.TestCheckResourceAttr(
					// 	"newrelicinfra_alert_condition.foo", "enabled", "false"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.duration", "10"),
				),
			},
			resource.TestStep{
				Config: testAccCheckNewRelicInfraAlertConditionConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicInfraAlertConditionExists("newrelicinfra_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "name", fmt.Sprintf("tf-test-updated-%s", rName)),
				),
			},
			resource.TestStep{
				Config: testAccCheckNewRelicInfraAlertConditionConfigWithWhere(rName, whereClause),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicInfraAlertConditionExists("newrelicinfra_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "where", whereClause),
				),
			},
		},
	})
}

func TestAccNewRelicInfraAlertCondition_Thresholds(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNewRelicInfraAlertConditionDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCheckNewRelicInfraAlertConditionConfigWithThreshold(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicInfraAlertConditionExists("newrelicinfra_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.duration", "10"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.value", "10"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.time_function", "any"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "warning.0.value", "20"),
				),
			},
			resource.TestStep{
				Config: testAccCheckNewRelicInfraAlertConditionConfigWithThresholdUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNewRelicInfraAlertConditionExists("newrelicinfra_alert_condition.foo"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "name", fmt.Sprintf("tf-test-%s", rName)),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.duration", "20"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.value", "15"),
					resource.TestCheckResourceAttr(
						"newrelicinfra_alert_condition.foo", "critical.0.time_function", "all"),
					resource.TestCheckNoResourceAttr(
						"newrelicinfra_alert_condition.foo", "warning"),
				),
			},
		},
	})
}

func testAccCheckNewRelicInfraAlertConditionDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*newrelic.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "newrelicinfra_alert_condition" {
			continue
		}

		ids, err := parseIDs(r.Primary.ID, 2)
		if err != nil {
			return err
		}

		policyID := ids[0]
		id := ids[1]

		_, err = client.GetAlertInfraCondition(policyID, id)
		if err == nil {
			return fmt.Errorf("Infra Alert condition still exists")
		}

	}
	return nil
}

func testAccCheckNewRelicInfraAlertConditionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No alert condition ID is set")
		}

		client := testAccProvider.Meta().(*newrelic.Client)

		ids, err := parseIDs(rs.Primary.ID, 2)
		if err != nil {
			return err
		}

		policyID := ids[0]
		id := ids[1]

		found, err := client.GetAlertInfraCondition(policyID, id)
		if err != nil {
			return err
		}

		if found.ID != id {
			return fmt.Errorf("Alert condition not found: %v - %v", id, found)
		}

		return nil
	}
}

func testAccCheckNewRelicInfraAlertConditionConfig(rName string) string {
	return fmt.Sprintf(`

resource "newrelicinfra_alert_condition" "foo" {
  policy_id = "211629"

  name            = "tf-test-%[1]s"
  # TODO: Still need to fix enabled 
  # enabled         = false

  type            = "infra_metric"
  event           = "StorageSample"
  select          = "diskFreePercent"
  comparison      = "below"

  critical {
	  duration = 10
	  value = 10
	  time_function = "any"
  }
}
`, rName)
}

func testAccCheckNewRelicInfraAlertConditionConfigWithWhere(rName, where string) string {
	return fmt.Sprintf(`

resource "newrelicinfra_alert_condition" "foo" {
  policy_id = "211629"

  name            = "tf-test-%[1]s"
  # TODO: Still need to fix enabled 
  # enabled         = false

  type            = "infra_metric"
  event           = "StorageSample"
  select          = "diskFreePercent"
  comparison      = "below"

  where = "%[2]s"

  critical {
	  duration = 10
	  value = 10
	  time_function = "any"
  }
}
`, rName, where)
}

func testAccCheckNewRelicInfraAlertConditionConfigUpdated(rName string) string {
	return fmt.Sprintf(`

resource "newrelicinfra_alert_condition" "foo" {
  policy_id = "211629"

  name            = "tf-test-updated-%[1]s"
  # TODO: Still need to fix enabled 
  # enabled         = false

  type            = "infra_metric"
  event           = "StorageSample"
  select          = "diskFreePercent"
  comparison      = "below"

  critical {
	  duration = 10
	  value = 10
	  time_function = "any"
  }
}
`, rName)
}

func testAccCheckNewRelicInfraAlertConditionConfigWithThreshold(rName string) string {
	return fmt.Sprintf(`

resource "newrelicinfra_alert_condition" "foo" {
  policy_id = "211629"

  name            = "tf-test-%[1]s"

  type            = "infra_metric"
  event           = "StorageSample"
  select          = "diskFreePercent"
  comparison      = "below"

  critical {
	duration = 10
	value = 10
	time_function = "any"
  }

  warning {
	duration = 10
	value = 20
	time_function = "any"
  }
}
`, rName)
}

func testAccCheckNewRelicInfraAlertConditionConfigWithThresholdUpdated(rName string) string {
	return fmt.Sprintf(`

resource "newrelicinfra_alert_condition" "foo" {
  policy_id = "211629"

  name            = "tf-test-%[1]s"

  type            = "infra_metric"
  event           = "StorageSample"
  select          = "diskFreePercent"
  comparison      = "below"

  critical {
    duration = 20
	value = 15
	time_function = "all"
  }
}
`, rName)
}
