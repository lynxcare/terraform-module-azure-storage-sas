package test

import (
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

	ctx := context.Background()

	container, blobEndpoint, err := setupStorageContainer(ctx, name, accountName, groupName, region, subscriptionID, clientID, clientSecret, tenantID)
	assert.NoError(t, err)
	assert.NotNil(t, container)

	defer func() {
		_, err = destroyResourceGroup(ctx, groupName, subscriptionID, clientID, clientSecret, tenantID)
		assert.NoError(t, err)
	}()

	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, "..", ".")

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
	_, err = terraform.InitAndApplyE(t, terraformOptions)
	assert.NoError(t, err)

	sasToken, err := terraform.OutputE(t, terraformOptions, "sas")
	assert.NoError(t, err)

	assert.NotEmpty(t, sasToken)

	err = tryBlobUpload(name, *blobEndpoint, sasToken, ctx)
	assert.NoError(t, err)

	_, err = terraform.ApplyE(t, terraformOptions)
	secondSasToken, err := terraform.OutputE(t, terraformOptions, "sas")

	assert.Equal(t, sasToken, secondSasToken)
	time.Sleep(3 * time.Minute)

	err = tryBlobUpload(name+"2", *blobEndpoint, sasToken, ctx)
	assert.NoError(t, err)

	_, err = terraform.ApplyE(t, terraformOptions)
	thirdSasToken, err := terraform.OutputE(t, terraformOptions, "sas")
	assert.NotEqual(t, sasToken, thirdSasToken)

	err = tryBlobUpload(name+"3", *blobEndpoint, thirdSasToken, ctx)
	assert.NoError(t, err)

	err = tryBlobUpload(name+"4", *blobEndpoint, sasToken, ctx)
	assert.Error(t, err)
}
