package tests

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/secrets-manager-go-sdk/v2/secretsmanagerv2"
	"github.com/gruntwork-io/terratest/modules/files"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/common"
	"github.com/terraform-ibm-modules/ibmcloud-terratest-wrapper/testprojects"
)

// Use existing resource group
const resourceGroup = "geretain-test-resources"

// Define a struct with fields that match the structure of the YAML data
const yamlLocation = "../common-dev-assets/common-go-assets/common-permanent-resources.yaml"

var permanentResources map[string]interface{}

// Current supported regions
var validRegions = []string{
	"us-south",
	"eu-de",
	"eu-es",
}

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

	acme_letsencrypt_private_key := GetSecretsManagerKey( // pragma: allowlist secret
		permanentResources["acme_letsencrypt_private_key_sm_id"].(string),
		permanentResources["acme_letsencrypt_private_key_sm_region"].(string),
		permanentResources["acme_letsencrypt_private_key_secret_id"].(string),
	)

	options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
		Testing:        t,
		Prefix:         "cs", // setting prefix here gets a random string appended to it
		ParallelDeploy: true,
	})

	options.SkipTestTearDown = true

	options.StackInputs = map[string]interface{}{
		"prefix":                               options.Prefix,
		"region":                               validRegions[rand.Intn(len(validRegions))],
		"existing_resource_group_name":         resourceGroup,
		"sm_service_plan":                      "standard",
		"secret_manager_iam_engine_enabled":    true,
		"secret_manager_public_engine_enabled": true,
		"existing_secrets_endpoint_type":       "private",
		"cis_id":                               permanentResources["cisInstanceId"],
		"ca_name":                              permanentResources["certificateAuthorityName"],
		"dns_provider_name":                    permanentResources["dnsProviderName"],
		"acme_letsencrypt_private_key":         *acme_letsencrypt_private_key,                              // pragma: allowlist secret
		"ibmcloud_api_key":                     options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], // always required by the stack
		"enable_platform_logs_metrics":         false,
		"en_email_list":                        []string{"GoldenEye.Operations@ibm.com"},
	}

	err := options.RunProjectsTest()
	if assert.NoError(t, err) {
		t.Log("TestProjectsFullTest Passed")
	} else {
		t.Error("TestProjectsFullTest Failed")
	}
}

func GetSecretsManagerKey(sm_id string, sm_region string, sm_key_id string) *string {
	secretsManagerService, err := secretsmanagerv2.NewSecretsManagerV2(&secretsmanagerv2.SecretsManagerV2Options{
		URL: fmt.Sprintf("https://%s.%s.secrets-manager.appdomain.cloud", sm_id, sm_region),
		Authenticator: &core.IamAuthenticator{
			ApiKey: os.Getenv("TF_VAR_ibmcloud_api_key"),
		},
	})
	if err != nil {
		panic(err)
	}

	getSecretOptions := secretsManagerService.NewGetSecretOptions(
		sm_key_id,
	)

	secret, _, err := secretsManagerService.GetSecret(getSecretOptions)
	if err != nil {
		panic(err)
	}
	return secret.(*secretsmanagerv2.ArbitrarySecret).Payload
}

func TestProjectsExistingResourcesTest(t *testing.T) {
	t.Parallel()

	// ------------------------------------------------------------------------------------
	// Provision RG, EN and SM
	// ------------------------------------------------------------------------------------

	region := validRegions[rand.Intn(len(validRegions))]
	prefix := fmt.Sprintf("css-ext-%s", strings.ToLower(random.UniqueId()))
	realTerraformDir := "./resources"
	tempTerraformDir, _ := files.CopyTerraformFolderToTemp(realTerraformDir, fmt.Sprintf(prefix+"-%s", strings.ToLower(random.UniqueId())))

	// Verify ibmcloud_api_key variable is set
	checkVariable := "TF_VAR_ibmcloud_api_key"
	val, present := os.LookupEnv(checkVariable)
	require.True(t, present, checkVariable+" environment variable not set")
	require.NotEqual(t, "", val, checkVariable+" environment variable is empty")
	logger.Log(t, "Tempdir: ", tempTerraformDir)
	existingTerraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTerraformDir,
		Vars: map[string]interface{}{
			"prefix": prefix,
			"region": region,
		},
		// Set Upgrade to true to ensure latest version of providers and modules are used by terratest.
		// This is the same as setting the -upgrade=true flag with terraform.
		Upgrade: true,
	})

	terraform.WorkspaceSelectOrNew(t, existingTerraformOptions, prefix)
	_, existErr := terraform.InitAndApplyE(t, existingTerraformOptions)
	if existErr != nil {
		assert.True(t, existErr == nil, "Init and Apply of temp existing resource failed")
	} else {

		// ------------------------------------------------------------------------------------
		// Test passing an existing SM, RG, EN
		// ------------------------------------------------------------------------------------

		options := testprojects.TestProjectOptionsDefault(&testprojects.TestProjectsOptions{
			Testing:        t,
			ParallelDeploy: true,
		})

		options.StackInputs = map[string]interface{}{
			"prefix":                            terraform.Output(t, existingTerraformOptions, "prefix"),
			"region":                            terraform.Output(t, existingTerraformOptions, "region"),
			"existing_resource_group_name":      terraform.Output(t, existingTerraformOptions, "resource_group_name"),
			"ibmcloud_api_key":                  options.RequiredEnvironmentVars["TF_VAR_ibmcloud_api_key"], // always required by the stack
			"enable_platform_logs_metrics":      false,
			"existing_secrets_manager_crn":      terraform.Output(t, existingTerraformOptions, "secrets_manager_instance_crn"),
			"secret_manager_iam_engine_enabled": true,
			"existing_kms_instance_crn":         permanentResources["hpcs_south_crn"],
			"en_email_list":                     []string{"GoldenEye.Operations@ibm.com"},
		}

		err := options.RunProjectsTest()
		if assert.NoError(t, err) {
			t.Log("TestProjectsExistingResourcesTest Passed")
		} else {
			t.Error("TestProjectsExistingResourcesTest Failed")
		}
	}

	// Check if "DO_NOT_DESTROY_ON_FAILURE" is set
	envVal, _ := os.LookupEnv("DO_NOT_DESTROY_ON_FAILURE")
	// Destroy the temporary existing resources if required
	if t.Failed() && strings.ToLower(envVal) == "true" {
		fmt.Println("Terratest failed. Debug the test and delete resources manually.")
	} else {
		logger.Log(t, "START: Destroy (existing resources)")
		terraform.Destroy(t, existingTerraformOptions)
		terraform.WorkspaceDelete(t, existingTerraformOptions, prefix)
		logger.Log(t, "END: Destroy (existing resources)")
	}
}
