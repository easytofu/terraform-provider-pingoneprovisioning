package client

import (
	"github.com/patrickcping/pingone-go-sdk-v2/management"
)

// Client holds the PingOne SDK clients shared by provider resources and datasources.
type Client struct {
	API *management.APIClient
}
