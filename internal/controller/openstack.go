package controller

import (
	"context"
	"fmt"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
)

// checkVMStatus checks the status of a VM by its ID and returns whether it's ACTIVE.
func checkVMStatus(ctx context.Context, computeClient *gophercloud.ServiceClient, vmID string) (bool, string, error) {
	// VM 정보 가져오기
	server, err := servers.Get(ctx, computeClient, vmID).Extract()
	if err != nil {
		return false, "", fmt.Errorf("failed to fetch VM details for ID %s: %w", vmID, err)
	}

	// 현재 상태 확인
	status := server.Status
	// 상태가 ACTIVE인지 확인
	isActive := status == "ACTIVE"
	return isActive, status, nil
}
