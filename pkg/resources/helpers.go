package resources

import (
	"fmt"

	"github.com/Snowflake-Labs/terraform-provider-snowflake/pkg/sdk"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func isOk(_ interface{}, ok bool) bool {
	return ok
}

func dataTypeValidateFunc(val interface{}, key string) (warns []string, errs []error) {
	if ok := sdk.IsValidDataType(val.(string)); !ok {
		errs = append(errs, fmt.Errorf("%v is not a valid data type", val))
	}
	return
}

func dataTypeDiffSuppressFunc(k, old, new string, d *schema.ResourceData) bool {
	old_dt, err := sdk.DataTypeFromString(old)
	if err != nil {
		return false
	}
	new_dt, err := sdk.DataTypeFromString(new)
	if err != nil {
		return false
	}
	return old_dt == new_dt
}
