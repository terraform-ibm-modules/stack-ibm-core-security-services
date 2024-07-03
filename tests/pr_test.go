package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testprojects"
)

// Use existing resource group
const resourceGroup = "geretain-test-resources"

func TestProjectsFullTest(t *testing.T) {

	options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
		Testing:        t,
		Prefix:         "cs", // setting prefix here gets a random string appended to it
		ParallelDeploy: true,
	})

	options.StackInputs = map[string]interface{}{
		"prefix":                            options.Prefix,
		"existing_resource_group_name":      resourceGroup,
		"sm_service_plan":                   "trial",
		"use_existing_resource_group":       false,
		"secret_manager_iam_engine_enabled": true,
		"ibmcloud_api_key":                  options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], // always required by the stack
		"enable_platform_logs_metrics":      false,
	}

	err := options.RunProjectsTest()
	if assert.NoError(t, err) {
		t.Log("TestProjectsFullTest Passed")
	} else {
		t.Error("TestProjectsFullTest Failed")
	}
}
