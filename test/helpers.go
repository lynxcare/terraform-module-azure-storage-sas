package test

import (
	"context"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

func tryBlobUpload(ctx context.Context, name string, blobEndpoint string, sasToken string) error {
	data := []byte("hello world this is a blob\n")
	fileName := "testblob_" + name
	err := ioutil.WriteFile(fileName, data, 0700)
	if err != nil {
		return err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	blobP := azblob.NewPipeline(azblob.NewAnonymousCredential(), azblob.PipelineOptions{})

	URL, err := url.Parse(blobEndpoint + name + "/" + fileName + sasToken)
	if err != nil {
		return err
	}

	blockBlobUrl := azblob.NewBlockBlobURL(*URL, blobP)
	_, err = azblob.UploadFileToBlockBlob(ctx, file, blockBlobUrl, azblob.UploadToBlockBlobOptions{
		BlockSize:   4 * 1024 * 1024,
		Parallelism: 16})
	return err
}

func setupStorageContainer(ctx context.Context, containerName string, storageAccountName string, resourceGroupName string, region string, subscriptionID string, clientID string, clientSecret string, tenantID string) (*storage.BlobContainer, *string, error) {
	_, err := createResourceGroup(ctx, resourceGroupName, region, subscriptionID, clientID, clientSecret, tenantID)
	if err != nil {
		return nil, nil, err
	}

	storageAccountClient, err := getStorageAccountsClient(subscriptionID, clientID, clientSecret, tenantID)
	if err != nil {
		return nil, nil, err
	}

	params := storage.AccountCreateParameters{
		Sku: &storage.Sku{
			Name: storage.StandardLRS,
		},
		Kind:     storage.StorageV2,
		Location: to.StringPtr(region),
		AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{
			EnableHTTPSTrafficOnly: to.BoolPtr(true),
		},
	}

	future, err := storageAccountClient.Create(ctx, resourceGroupName, storageAccountName, params)
	if err != nil {
		return nil, nil, err
	}

	err = future.WaitForCompletionRef(ctx, storageAccountClient.Client)
	if err != nil {
		return nil, nil, err
	}

	acc, err := future.Result(*storageAccountClient)
	if err != nil {
		return nil, nil, err
	}

	blobEndpoint := acc.PrimaryEndpoints.Blob

	containerClient, err := getStorageContainerClient(subscriptionID, clientID, clientSecret, tenantID)
	if err != nil {
		return nil, nil, err
	}

	container, err := containerClient.Create(ctx, resourceGroupName, storageAccountName, containerName, storage.BlobContainer{})
	return &container, blobEndpoint, err
}

func getStorageContainerClient(subscriptionID string, clientID string, clientSecret string, tenantID string) (*storage.BlobContainersClient, error) {
	containerClient := storage.NewBlobContainersClient(subscriptionID)
	authorizer, err := getARMAuthorizer(clientID, clientSecret, tenantID)
	if err != nil {
		return nil, err
	}

	containerClient.Authorizer = authorizer
	return &containerClient, err
}

func getStorageAccountsClient(subscriptionID string, clientID string, clientSecret string, tenantID string) (*storage.AccountsClient, error) {
	accountsClient := storage.NewAccountsClient(subscriptionID)
	authorizer, err := getARMAuthorizer(clientID, clientSecret, tenantID)
	if err != nil {
		return nil, err
	}

	accountsClient.Authorizer = authorizer
	return &accountsClient, err
}

func createResourceGroup(ctx context.Context, name string, region string, subscriptionID string, clientID string, clientSecret string, tenantID string) (*resources.Group, error) {
	rgClient, err := getResourceGroupsClient(subscriptionID, clientID, clientSecret, tenantID)
	if err != nil {
		return nil, err
	}

	rg, err := rgClient.CreateOrUpdate(ctx, name, resources.Group{
		Location: to.StringPtr(region),
	})
	return &rg, err
}

func destroyResourceGroup(ctx context.Context, resourceGroupName string, subscriptionID string, clientID string, clientSecret string, tenantID string) (*resources.GroupsDeleteFuture, error) {
	rgClient, err := getResourceGroupsClient(subscriptionID, clientID, clientSecret, tenantID)
	if err != nil {
		return nil, err
	}

	future, err := rgClient.Delete(ctx, resourceGroupName)
	return &future, err
}

func getResourceGroupsClient(subscriptionID string, clientID string, clientSecret string, tenantID string) (*resources.GroupsClient, error) {
	rgClient := resources.NewGroupsClient(subscriptionID)
	authorizer, err := getARMAuthorizer(clientID, clientSecret, tenantID)
	if err != nil {
		return nil, err
	}

	rgClient.Authorizer = authorizer
	return &rgClient, err
}

func getARMAuthorizer(clientID string, clientSecret string, tenantID string) (autorest.Authorizer, error) {
	authorizer, err := auth.NewClientCredentialsConfig(clientID, clientSecret, tenantID).Authorizer()
	return authorizer, err
}
