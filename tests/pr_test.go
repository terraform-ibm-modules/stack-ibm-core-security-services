package tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testprojects"
	"os"
	"testing"
)

func TestProjectsFullTest(t *testing.T) {

	options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
		Testing: t,
		StackConfigurationOrder: []string{
			"primary-da",
			"secondary-da",
		},
	})
	options.StackMemberInputs = map[string]map[string]interface{}{
		"primary-da": {
			"prefix": fmt.Sprintf("p%s", options.Prefix),
		},
		"secondary-da": {
			"prefix": fmt.Sprintf("s%s", options.Prefix),
		},
	}
	options.StackInputs = map[string]interface{}{
		"resource_group_name": options.ResourceGroup,
		"ibmcloud_api_key":    os.Getenv("TF_VAR_ibmcloud_api_key"),
	}

	err := options.RunProjectsTest()
	if assert.NoError(t, err) {
		t.Log("TestProjectsFullTest Passed")
	} else {
		t.Error("TestProjectsFullTest Failed")
	}
}
