package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/pingidentity/terraform-provider-pingoneprovisioning/internal/provider"
)

var version = "dev"

func main() {
	providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.opentofu.org/pingidentity/pingoneprovisioning",
	})
}
