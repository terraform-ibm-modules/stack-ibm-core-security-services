package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testprojects"
	"testing"
)

func TestProjectsFullTest(t *testing.T) {

	options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
		Testing:        t,
		Prefix:         "cs", // setting prefix here gets a random string appended to it
		ParallelDeploy: true,
	})

	options.StackMemberInputs = map[string]map[string]interface{}{
		"3 - core-security-observability": {
			"enable_platform_logs": false,
		},
	}
	options.StackInputs = map[string]interface{}{
		"prefix":                      options.Prefix,
		"resource_group_name":         options.Prefix,
		"sm_service_plan":             "trial",
		"use_existing_resource_group": false,
		"ibmcloud_api_key":            options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], // always required by the stack
	}

	err := options.RunProjectsTest()
	if assert.NoError(t, err) {
		t.Log("TestProjectsFullTest Passed")
	} else {
		t.Error("TestProjectsFullTest Failed")
	}
}
