package resources_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAcc_Account_complete(t *testing.T) {
	// SNOWFLAKE_TEST_ACCOUNT_CREATE must be set to 1 to run this test
	if _, ok := os.LookupEnv("SNOWFLAKE_TEST_ACCOUNT_CREATE"); !ok {
		t.Skip("Skipping TestInt_AccountCreate")
	}
	accountName := strings.ToUpper(acctest.RandStringFromCharSet(10, acctest.CharSetAlpha))
	password := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha) + "123ABC"

	resource.ParallelTest(t, resource.TestCase{
		Providers:    providers(),
		CheckDestroy: nil,
		// this errors with: Error running post-test destroy, there may be dangling resources: exit status 1
		// unless we change the resource to return nil on destroy then this is unavoidable
		Steps: []resource.TestStep{
			{
				Config: accountConfig(accountName, password, "Terraform acceptance test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_account.test", "name", accountName),
					resource.TestCheckResourceAttr("snowflake_account.test", "admin_name", "someadmin"),
					resource.TestCheckResourceAttr("snowflake_account.test", "first_name", "Ad"),
					resource.TestCheckResourceAttr("snowflake_account.test", "last_name", "Min"),
					resource.TestCheckResourceAttr("snowflake_account.test", "email", "admin@example.com"),
					resource.TestCheckResourceAttr("snowflake_account.test", "must_change_password", "false"),
					resource.TestCheckResourceAttr("snowflake_account.test", "edition", "BUSINESS_CRITICAL"),
					resource.TestCheckResourceAttr("snowflake_account.test", "comment", "Terraform acceptance test"),
				),
				Destroy: false,
			},
			// UPDATE COMMENT
			{
				Config: accountConfig(accountName, password, "Please delete me!"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("snowflake_account.test", "name", accountName),
					resource.TestCheckResourceAttr("snowflake_account.test", "comment", "Please delete me!"),
				),

				Destroy:     false,
				ExpectError: regexp.MustCompile("Error: cannot delete Snowflake accounts(.*)"),
			},
			// IMPORT
			{
				ResourceName:      "snowflake_account.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"admin_name",
					"admin_password",
					"admin_rsa_public_key",
					"email",
					"must_change_password",
					"first_name",
					"last_name",
				},
				Destroy: false,
			},
		},
	})
}

func accountConfig(name string, password string, comment string) string {
	return fmt.Sprintf(`
data "snowflake_current_account" "current" {}

resource "snowflake_account" "test" {
  name = "%s"
  admin_name = "someadmin"
  admin_password = "%s"
  first_name = "Ad"
  last_name = "Min"
  email = "admin@example.com"
  must_change_password = false
  edition = "BUSINESS_CRITICAL"
  comment = "%s"
  region = data.snowflake_current_account.current.region
}
`, name, password, comment)
}