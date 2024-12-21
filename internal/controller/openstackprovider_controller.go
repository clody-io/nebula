/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	virtualenvv1 "github.com/clody-io/nebula/api/v1"
	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack"
	"github.com/gophercloud/gophercloud/v2/openstack/compute/v2/keypairs"
	instance "github.com/gophercloud/gophercloud/v2/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/extensions/layer3/floatingips"
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/ports"
	logger "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"log"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OpenstackProviderReconciler reconciles a OpenstackProvider object
type OpenstackProviderReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	OpenstackClient *gophercloud.ProviderClient
}

// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackvms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackvms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackvms/finalizers,verbs=update
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackproviders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackproviders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=virtual-env.clody.io,resources=openstackproviders/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the OpenstackProvider object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile

func (r *OpenstackProviderReconciler) createOrUpdateOpenstackVM(ctx context.Context,
	openstackVM *virtualenvv1.OpenstackVM) (ctrl.Result, error) {

	namespacedName := types.NamespacedName{
		Namespace: openstackVM.Namespace,
		Name:      openstackVM.Name,
	}

	logCtx := logger.WithField("OpenstackVMController", namespacedName)

	computeClient, err := openstack.NewComputeV2(r.OpenstackClient, gophercloud.EndpointOpts{
		Region: "kr-central-01", // 사용할 OpenStack 리전 이름
	})
	if err != nil {
		panic(fmt.Errorf("failed to create Compute client: %w", err))
	}

	if openstackVM.Status.VirtualMachineStatus.VirtualMachine == "" {

		listOpts := instance.ListOpts{
			Name: openstackVM.Name,
		}

		allPages, err := instance.List(computeClient, listOpts).AllPages(ctx)
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, fmt.Errorf("failed to list servers: %v", err)
		}

		allServers, err := instance.ExtractServers(allPages)
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, fmt.Errorf("failed to extract servers: %v", err)
		}

		// 동일한 이름의 서버가 있는지 확인
		if len(allServers) > 0 {
			logCtx.Printf("Server with name %s already exists: %s\n", openstackVM.Name, allServers[0].ID)
			openstackVM.Status.InstanceStatus.InstanceID = allServers[0].ID
			openstackVM.Status.VirtualMachineStatus.VirtualMachine = allServers[0].ID
			openstackVM.Status.VirtualMachineStatus.Status = allServers[0].Status
			openstackVM.Status.VirtualMachineStatus.Step = "Launching VM"
			openstackVM.Status.VirtualMachineStatus.LastTransitionTime = &metav1.Time{Time: time.Now()}

			if err := r.Status().Update(ctx, openstackVM); err != nil {
				return ctrl.Result{RequeueAfter: 10 * time.Second},
					fmt.Errorf("failed to update OpenstackVM status: %w", err)
			}

			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}

		createOpts := instance.CreateOpts{
			Name:             openstackVM.Name, // 서버 이름
			FlavorRef:        "0db5a869-052f-49e6-adbe-e5c12baddbf8",
			ImageRef:         "022db2f8-1af1-4bc8-a8e7-c14fa1332389", // Image ID
			AvailabilityZone: "nova",
			SecurityGroups:   []string{"default"},
			AdminPass:        "Tmaltmfhdn1!",
			Networks:         []instance.Network{{UUID: "6520413c-b52b-4931-8804-0e8b834e12d7"}}, // 네트워크 ID}, // 볼륨 옵션 추가
		}
		//CreateOptsBuilder, hintOpts SchedulerHintOptsBuilder

		server, err := instance.Create(ctx, computeClient, keypairs.CreateOptsExt{
			CreateOptsBuilder: createOpts,
			KeyName:           "clody-infra",
		}, instance.SchedulerHintOpts{}).Extract()
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second}, fmt.Errorf("failed to create VM: %w", err)
		}

		// 상태 업데이트
		openstackVM.Status.InstanceStatus.InstanceID = server.ID
		openstackVM.Status.VirtualMachineStatus.VirtualMachine = server.ID
		openstackVM.Status.VirtualMachineStatus.Status = "BUILD"
		openstackVM.Status.VirtualMachineStatus.Step = "Launching VM"
		openstackVM.Status.VirtualMachineStatus.LastTransitionTime = &metav1.Time{Time: time.Now()}

		if err := r.Status().Update(ctx, openstackVM); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("failed to update OpenstackVM status: %w", err)
		}

		logCtx.Printf("VM creation initiated for OpenstackVM: %s, VM ID: %s", openstackVM.Name, server.ID)
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	} else {
		logCtx.Println("Check VM Status..")

		isActive, _, err := checkVMStatus(ctx, computeClient, openstackVM.Status.VirtualMachineStatus.VirtualMachine)
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("failed to update OpenstackVM status: %w", err)
		}
		if !isActive {
			logCtx.Println("Will Attempt after 10 seconds..")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}

		// 이미 Floating IP가 연결된 경우 동작 중단
		if openstackVM.Status.InstanceStatus.FloatingIP != "" {
			logCtx.Printf("Floating IP already associated: %s. Skipping further actions.",
				openstackVM.Status.InstanceStatus.FloatingIP)
			return ctrl.Result{}, nil
		}

		// Network Client 생성
		networkClient, err := openstack.NewNetworkV2(r.OpenstackClient, gophercloud.EndpointOpts{
			Region: "kr-central-01",
		})
		if err != nil {
			log.Fatalf("Failed to create Network client: %v", err)
		}

		// Port 확인하고 floting ip가 붙어있는지 확인

		var portID string
		logCtx.Infof("Checking if Openstack VM has floating ip...")
		portPages, err := ports.List(networkClient, ports.ListOpts{
			DeviceID: openstackVM.Status.VirtualMachineStatus.VirtualMachine, // 서버와 연결된 포트 필터링
		}).AllPages(ctx)
		if err != nil {
			logCtx.Fatalf("Failed to list ports: %v", err)
		}
		ports, err := ports.ExtractPorts(portPages)
		if err != nil {
			logCtx.Fatalf("Failed to extract ports: %v", err)
		}
		if len(ports) > 0 {
			portID = ports[0].ID // 첫 번째 포트를 선택
		}
		if portID == "" {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("No ports found for server ID: %s",
					openstackVM.Status.VirtualMachineStatus.VirtualMachine)
		}

		fipPages, err := floatingips.List(networkClient, floatingips.ListOpts{
			PortID: portID, // 포트 ID에 연결된 Floating IP 필터링
		}).AllPages(context.Background())
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("failed to list floating IPs: %v", err)
		}

		fipList, err := floatingips.ExtractFloatingIPs(fipPages)
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("failed to extract floating IPs: %v", err)
		}

		if len(fipList) > 0 {
			logCtx.Infof("Openstack VM already has floating ip")
			openstackVM.Status.FloatingIP = fipList[0].FloatingIP
			openstackVM.Status.VirtualMachineStatus.LastTransitionTime = &metav1.Time{Time: time.Now()}
			if err := r.Status().Update(ctx, openstackVM); err != nil {
				return ctrl.Result{RequeueAfter: 10 * time.Second},
					fmt.Errorf("failed to update OpenstackVM status: %w", err)
			}
			return ctrl.Result{}, nil
		}

		createOpts := floatingips.CreateOpts{
			FloatingNetworkID: "84f3204e-87bf-4321-9814-6f9fa747b7b2", // External 네트워크 ID
			SubnetID:          "24a1bf72-1586-439d-98c4-fd4821fc0c3f",
		}
		fip, err := floatingips.Create(ctx, networkClient, createOpts).Extract()
		if err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("failed to update OpenstackVM status: %w", err)
		}

		logCtx.Printf("Created Floating IP: %s\n", fip.FloatingIP)

		// 4. Floating IP를 인스턴스에 연결
		updateOpts := floatingips.UpdateOpts{
			PortID: &portID, // 연결할 포트 ID
		}
		_, err = floatingips.Update(ctx, networkClient, fip.ID, updateOpts).Extract()
		if err != nil {
			logCtx.Fatalf("Failed to associate Floating IP: %v", err)
		}

		openstackVM.Status.VirtualMachineStatus.Status = "ACTIVE"
		openstackVM.Status.FloatingIP = fip.FloatingIP
		openstackVM.Status.VirtualMachineStatus.LastTransitionTime = &metav1.Time{Time: time.Now()}
		if err := r.Status().Update(ctx, openstackVM); err != nil {
			return ctrl.Result{RequeueAfter: 10 * time.Second},
				fmt.Errorf("failed to update OpenstackVM status: %w", err)
		}
		return ctrl.Result{}, nil
	}
}

func (r *OpenstackProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here
	openstackVM := &virtualenvv1.OpenstackVM{}
	if err := r.Get(ctx, req.NamespacedName, openstackVM); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return r.createOrUpdateOpenstackVM(ctx, openstackVM)
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenstackProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&virtualenvv1.OpenstackVM{}).
		Complete(r)
}
