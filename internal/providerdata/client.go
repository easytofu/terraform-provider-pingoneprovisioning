package providerdata

import (
	"github.com/patrickcping/pingone-go-sdk-v2/management"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/githubapi"
)

// Client holds the PingOne SDK clients shared by provider resources and datasources.
type Client struct {
	API    *management.APIClient
	GitHub *githubapi.Client
}
