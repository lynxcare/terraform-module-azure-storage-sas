package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/random"

	"context"

	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

func TestRotationEvery3MinutesWith2MinuteMargin(t *testing.T) {
	t.Parallel()

	subscriptionID, exists := os.LookupEnv("ARM_SUBSCRIPTION_ID")
	assert.True(t, exists)

	clientID, exists := os.LookupEnv("ARM_CLIENT_ID")
	assert.True(t, exists)

	clientSecret, exists := os.LookupEnv("ARM_CLIENT_SECRET")
	assert.True(t, exists)

	tenantID, exists := os.LookupEnv("ARM_TENANT_ID")
	assert.True(t, exists)

	region := "eastus2"
	name := strings.ToLower(random.UniqueId())
	accountName := "sa" + name
	groupName := "rg" + name
	containerName := "c" + name
	fileName := "testblob"

	ctx := context.Background()

	t.Logf("Creating storage account %s & container %s", accountName, containerName)
	container, blobEndpoint, err := setupStorageContainer(ctx, containerName, accountName, groupName, region, subscriptionID, clientID, clientSecret, tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, container)
	t.Logf("Created storage account %s and container %s", accountName, containerName)

	defer func() {
		t.Log("Destroying resource group")
		_, err = destroyResourceGroup(ctx, groupName, subscriptionID, clientID, clientSecret, tenantID)
		assert.NoError(t, err)
	}()

	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, "..", ".")
	files, err := ioutil.ReadDir(tempTestFolder)
	assert.NoError(t, err)

	for _, f := range files {
		fileName := f.Name()
		if strings.Contains(fileName, ".tf.tests") {
			err := os.Rename(fmt.Sprintf("%s/%s", tempTestFolder, fileName), fmt.Sprintf("%s/%s", tempTestFolder, strings.Replace(fileName, ".tf.tests", ".tf", 1)))
			assert.NoError(t, err)
		}
	}

	terraformOptions := &terraform.Options{
		TerraformDir: tempTestFolder,
		Vars:         map[string]interface{}{},
		NoColor:      true,
		Logger:       logger.TestingT,
	}

	terraformOptions.Vars["rotation_minutes"] = 3
	terraformOptions.Vars["rotation_margin"] = "2m"
	terraformOptions.Vars["resource_group_name"] = groupName
	terraformOptions.Vars["storage_account_name"] = accountName

	defer terraform.Destroy(t, terraformOptions)
	t.Log("Running Terraform apply")
	_, err = terraform.InitAndApplyE(t, terraformOptions)
	assert.NoError(t, err)

	t.Log("Getting first SAS token")
	sasToken, err := terraform.OutputE(t, terraformOptions, "sas")
	assert.NoError(t, err)

	assert.NotEmpty(t, sasToken)

	t.Log("Trying blob upload with first SAS token, should succeed")
	err = tryBlobUpload(ctx, fileName, containerName, *blobEndpoint, sasToken)
	assert.NoError(t, err)

	t.Log("Running Terraform apply again")
	_, err = terraform.ApplyE(t, terraformOptions)
	assert.NoError(t, err)

	t.Log("Getting second SAS token, should not be different from first")
	secondSasToken, err := terraform.OutputE(t, terraformOptions, "sas")
	assert.NoError(t, err)
	assert.Equal(t, sasToken, secondSasToken)

	t.Log("Sleep for 3 minutes")
	time.Sleep(3 * time.Minute)

	t.Log("Trying blob upload with first SAS token, should still work")
	err = tryBlobUpload(ctx, fileName+"2", containerName, *blobEndpoint, sasToken)
	assert.NoError(t, err)

	t.Log("Running Terraform apply again")
	_, err = terraform.ApplyE(t, terraformOptions)
	assert.NoError(t, err)

	t.Log("Getting third SAS token, should be different from first")
	thirdSasToken, err := terraform.OutputE(t, terraformOptions, "sas")
	assert.NotEqual(t, sasToken, thirdSasToken)

	t.Log("Trying blob upload with third SAS token, should succeed")
	err = tryBlobUpload(ctx, fileName+"3", containerName, *blobEndpoint, thirdSasToken)
	assert.NoError(t, err)

	t.Log("Sleep for 2 minutes")
	time.Sleep(2 * time.Minute)

	t.Log("Trying blob upload with first SAS token, should not work anymore")
	err = tryBlobUpload(ctx, fileName+"4", containerName, *blobEndpoint, sasToken)
	assert.Error(t, err)
}
