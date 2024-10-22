package util

import (
	"fmt"
	"os"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

type OpenStackClient struct {
	provider       *gophercloud.ProviderClient
	identityClient *gophercloud.ServiceClient
	networkClient  *gophercloud.ServiceClient
	computeClient  *gophercloud.ServiceClient
}

func NewOpenStackClient(authURL string) *OpenStackClient {
	var err error

	providerClient, err := auth()

	if err != nil {
		// TODO. Change log format.
		fmt.Fprintf(os.Stderr, "OpenStack Authentication provider ERROR : %v \n", err)
	}

	identityClient, err := openstack.NewIdentityV3(providerClient, gophercloud.EndpointOpts{Region: os.Getenv("OS_REGION_NAME")})
	if err != nil {
		// TODO. Change log format.
		fmt.Fprintf(os.Stderr, "OpenStack Authentication identity ERROR : %v \n", err)
	}

	networkClient, err := openstack.NewNetworkV2(providerClient, gophercloud.EndpointOpts{Region: os.Getenv("OS_REGION_NAME")})
	if err != nil {
		// TODO. Change log format.
		fmt.Fprintf(os.Stderr, "OpenStack Authentication network ERROR : %v \n", err)
	}

	computeClient, err := openstack.NewComputeV2(providerClient, gophercloud.EndpointOpts{Region: os.Getenv("OS_REGION_NAME")})
	if err != nil {
		// TODO. Change log format.
		fmt.Fprintf(os.Stderr, "OpenStack Authentication network ERROR : %v \n", err)
	}

	return &OpenStackClient{
		provider:       providerClient,
		identityClient: identityClient,
		networkClient:  networkClient,
		computeClient:  computeClient,
	}
}

func auth() (*gophercloud.ProviderClient, error) {
	opts, err := openstack.AuthOptionsFromEnv()
	// setting reauth value true.
	// https://github.com/gophercloud/gophercloud/blob/master/auth_options.go#L70-L74
	opts.AllowReauth = true
	if err != nil {
		return nil, err
	}
	return openstack.AuthenticatedClient(opts)
}
