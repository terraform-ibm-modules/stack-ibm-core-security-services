package tests

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/common"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testprojects"
)

// Use existing resource group
const resourceGroup = "geretain-test-resources"

// Define a struct with fields that match the structure of the YAML data
const yamlLocation = "../common-dev-assets/common-go-assets/common-permanent-resources.yaml"

var permanentResources map[string]interface{}

func TestMain(m *testing.M) {
	// Read the YAML file contents
	var err error
	permanentResources, err = common.LoadMapFromYaml(yamlLocation)
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestProjectsFullTest(t *testing.T) {
	t.Parallel()
	options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
		Testing:        t,
		Prefix:         "cs", // setting prefix here gets a random string appended to it
		ParallelDeploy: true,
	})

	options.StackInputs = map[string]interface{}{
		"prefix":                            options.Prefix,
		"existing_resource_group_name":      resourceGroup,
		"sm_service_plan":                   "trial",
		"secret_manager_iam_engine_enabled": true,
		"ibmcloud_api_key":                  options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], // always required by the stack
		"enable_platform_logs_metrics":      false,
		"en_email_list":                     []string{"GoldenEye.Operations@ibm.com"},
	}

	err := options.RunProjectsTest()
	if assert.NoError(t, err) {
		t.Log("TestProjectsFullTest Passed")
	} else {
		t.Error("TestProjectsFullTest Failed")
	}
}

func TestProjectsExistingResourcesTest(t *testing.T) {
	t.Parallel()
	options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
		Testing:        t,
		Prefix:         "ecs", // setting prefix here gets a random string appended to it
		ParallelDeploy: true,
	})

	options.StackInputs = map[string]interface{}{
		"prefix":                       options.Prefix,
		"existing_resource_group_name": resourceGroup,
		"ibmcloud_api_key":             options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], // always required by the stack
		"enable_platform_logs_metrics": false,
		// More info: https://github.ibm.com/GoldenEye/issues/issues/9709#issuecomment-83874969
		// "existing_secrets_manager_crn": permanentResources["secretsManagerCRN"],
		"sm_service_plan":           "trial",
		"existing_kms_instance_crn": permanentResources["hpcs_south_crn"],
		"en_email_list":             []string{"GoldenEye.Operations@ibm.com"},
	}

	err := options.RunProjectsTest()
	if assert.NoError(t, err) {
		t.Log("TestProjectsExistingResourcesTest Passed")
	} else {
		t.Error("TestProjectsExistingResourcesTest Failed")
	}
}
