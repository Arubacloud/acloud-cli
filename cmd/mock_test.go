package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ─── top-level client mock ────────────────────────────────────────────────────

type mockClient struct {
	computeClient   aruba.ComputeClient
	networkClient   aruba.NetworkClient
	storageClient   aruba.StorageClient
	databaseClient  aruba.DatabaseClient
	projectClient   aruba.ProjectClient
	scheduleClient  aruba.ScheduleClient
	containerClient aruba.ContainerClient
	securityClient  aruba.SecurityClient
	auditClient     aruba.AuditClient
	metricClient    aruba.MetricClient
}

func (m *mockClient) FromCompute() aruba.ComputeClient   { return m.computeClient }
func (m *mockClient) FromNetwork() aruba.NetworkClient   { return m.networkClient }
func (m *mockClient) FromStorage() aruba.StorageClient   { return m.storageClient }
func (m *mockClient) FromDatabase() aruba.DatabaseClient { return m.databaseClient }
func (m *mockClient) FromProject() aruba.ProjectClient   { return m.projectClient }
func (m *mockClient) FromSchedule() aruba.ScheduleClient { return m.scheduleClient }
func (m *mockClient) FromContainer() aruba.ContainerClient {
	return m.containerClient
}
func (m *mockClient) FromSecurity() aruba.SecurityClient { return m.securityClient }
func (m *mockClient) FromAudit() aruba.AuditClient       { return m.auditClient }
func (m *mockClient) FromMetric() aruba.MetricClient     { return m.metricClient }

// ─── compute ─────────────────────────────────────────────────────────────────

type mockComputeClient struct {
	cloudServersClient aruba.CloudServersClient
	keyPairsClient     aruba.KeyPairsClient
}

func (m *mockComputeClient) CloudServers() aruba.CloudServersClient { return m.cloudServersClient }
func (m *mockComputeClient) KeyPairs() aruba.KeyPairsClient         { return m.keyPairsClient }

type mockCloudServersClient struct {
	listFn        func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.CloudServerList], error)
	getFn         func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	createFn      func(ctx context.Context, projectID string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	updateFn      func(ctx context.Context, projectID string, id string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	deleteFn      func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[any], error)
	powerOnFn     func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	powerOffFn    func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error)
	setPasswordFn func(ctx context.Context, projectID string, id string, body types.CloudServerPasswordRequest, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockCloudServersClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.CloudServerList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.CloudServerList]{StatusCode: 200, Data: &types.CloudServerList{}}, nil
}
func (m *mockCloudServersClient) Get(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, id, params)
	}
	return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
}
func (m *mockCloudServersClient) Create(ctx context.Context, projectID string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
}
func (m *mockCloudServersClient) Update(ctx context.Context, projectID, id string, body types.CloudServerRequest, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, id, body, params)
	}
	return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
}
func (m *mockCloudServersClient) Delete(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, id, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}
func (m *mockCloudServersClient) PowerOn(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	if m.powerOnFn != nil {
		return m.powerOnFn(ctx, projectID, id, params)
	}
	return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
}
func (m *mockCloudServersClient) PowerOff(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.CloudServerResponse], error) {
	if m.powerOffFn != nil {
		return m.powerOffFn(ctx, projectID, id, params)
	}
	return &types.Response[types.CloudServerResponse]{StatusCode: 200, Data: &types.CloudServerResponse{}}, nil
}
func (m *mockCloudServersClient) SetPassword(ctx context.Context, projectID, id string, body types.CloudServerPasswordRequest, params *types.RequestParameters) (*types.Response[any], error) {
	if m.setPasswordFn != nil {
		return m.setPasswordFn(ctx, projectID, id, body, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockKeyPairsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error)
	getFn    func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error)
	createFn func(ctx context.Context, projectID string, body types.KeyPairRequest, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error)
	deleteFn func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockKeyPairsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.KeyPairListResponse]{StatusCode: 200, Data: &types.KeyPairListResponse{}}, nil
}
func (m *mockKeyPairsClient) Get(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, id, params)
	}
	return &types.Response[types.KeyPairResponse]{StatusCode: 200, Data: &types.KeyPairResponse{}}, nil
}
func (m *mockKeyPairsClient) Create(ctx context.Context, projectID string, body types.KeyPairRequest, params *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.KeyPairResponse]{StatusCode: 200, Data: &types.KeyPairResponse{}}, nil
}
func (m *mockKeyPairsClient) Delete(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, id, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── network ─────────────────────────────────────────────────────────────────

type mockNetworkClient struct {
	vpcsMock             *mockVPCsClient
	elasticIPsMock       *mockElasticIPsClient
	loadBalancersMock    aruba.LoadBalancersClient
	securityGroupsMock   aruba.SecurityGroupsClient
	securityGroupRules   aruba.SecurityGroupRulesClient
	subnetsMock          aruba.SubnetsClient
	vpcPeeringsMock      aruba.VPCPeeringsClient
	vpcPeeringRoutesMock aruba.VPCPeeringRoutesClient
	vpnTunnelsMock       aruba.VPNTunnelsClient
	vpnRoutesMock        aruba.VPNRoutesClient
}

func (m *mockNetworkClient) VPCs() aruba.VPCsClient { return m.vpcsMock }
func (m *mockNetworkClient) ElasticIPs() aruba.ElasticIPsClient {
	return m.elasticIPsMock
}
func (m *mockNetworkClient) LoadBalancers() aruba.LoadBalancersClient {
	return m.loadBalancersMock
}
func (m *mockNetworkClient) SecurityGroups() aruba.SecurityGroupsClient {
	return m.securityGroupsMock
}
func (m *mockNetworkClient) SecurityGroupRules() aruba.SecurityGroupRulesClient {
	return m.securityGroupRules
}
func (m *mockNetworkClient) Subnets() aruba.SubnetsClient         { return m.subnetsMock }
func (m *mockNetworkClient) VPCPeerings() aruba.VPCPeeringsClient { return m.vpcPeeringsMock }
func (m *mockNetworkClient) VPCPeeringRoutes() aruba.VPCPeeringRoutesClient {
	return m.vpcPeeringRoutesMock
}
func (m *mockNetworkClient) VPNTunnels() aruba.VPNTunnelsClient { return m.vpnTunnelsMock }
func (m *mockNetworkClient) VPNRoutes() aruba.VPNRoutesClient   { return m.vpnRoutesMock }

type mockVPCsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPCList], error)
	getFn    func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.VPCResponse], error)
	createFn func(ctx context.Context, projectID string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error)
	updateFn func(ctx context.Context, projectID string, id string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error)
	deleteFn func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockVPCsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPCList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.VPCList]{StatusCode: 200, Data: &types.VPCList{}}, nil
}
func (m *mockVPCsClient) Get(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, id, params)
	}
	return &types.Response[types.VPCResponse]{StatusCode: 200, Data: &types.VPCResponse{}}, nil
}
func (m *mockVPCsClient) Create(ctx context.Context, projectID string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.VPCResponse]{StatusCode: 200, Data: &types.VPCResponse{}}, nil
}
func (m *mockVPCsClient) Update(ctx context.Context, projectID, id string, body types.VPCRequest, params *types.RequestParameters) (*types.Response[types.VPCResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, id, body, params)
	}
	return &types.Response[types.VPCResponse]{StatusCode: 200, Data: &types.VPCResponse{}}, nil
}
func (m *mockVPCsClient) Delete(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, id, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockElasticIPsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ElasticList], error)
	getFn    func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error)
	createFn func(ctx context.Context, projectID string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error)
	updateFn func(ctx context.Context, projectID string, id string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error)
	deleteFn func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockElasticIPsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ElasticList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.ElasticList]{StatusCode: 200, Data: &types.ElasticList{}}, nil
}
func (m *mockElasticIPsClient) Get(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, id, params)
	}
	return &types.Response[types.ElasticIPResponse]{StatusCode: 200, Data: &types.ElasticIPResponse{}}, nil
}
func (m *mockElasticIPsClient) Create(ctx context.Context, projectID string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.ElasticIPResponse]{StatusCode: 200, Data: &types.ElasticIPResponse{}}, nil
}
func (m *mockElasticIPsClient) Update(ctx context.Context, projectID, id string, body types.ElasticIPRequest, params *types.RequestParameters) (*types.Response[types.ElasticIPResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, id, body, params)
	}
	return &types.Response[types.ElasticIPResponse]{StatusCode: 200, Data: &types.ElasticIPResponse{}}, nil
}
func (m *mockElasticIPsClient) Delete(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, id, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── storage ──────────────────────────────────────────────────────────────────

type mockStorageClient struct {
	volumesMock   *mockVolumesClient
	snapshotsMock aruba.SnapshotsClient
	backupsMock   aruba.StorageBackupsClient
	restoresMock  aruba.StorageRestoreClient
}

func (m *mockStorageClient) Volumes() aruba.VolumesClient         { return m.volumesMock }
func (m *mockStorageClient) Snapshots() aruba.SnapshotsClient     { return m.snapshotsMock }
func (m *mockStorageClient) Backups() aruba.StorageBackupsClient  { return m.backupsMock }
func (m *mockStorageClient) Restores() aruba.StorageRestoreClient { return m.restoresMock }

type mockVolumesClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.BlockStorageList], error)
	getFn    func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error)
	createFn func(ctx context.Context, projectID string, body types.BlockStorageRequest, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error)
	updateFn func(ctx context.Context, projectID string, id string, body types.BlockStorageRequest, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error)
	deleteFn func(ctx context.Context, projectID string, id string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockVolumesClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.BlockStorageList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.BlockStorageList]{StatusCode: 200, Data: &types.BlockStorageList{}}, nil
}
func (m *mockVolumesClient) Get(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, id, params)
	}
	return &types.Response[types.BlockStorageResponse]{StatusCode: 200, Data: &types.BlockStorageResponse{}}, nil
}
func (m *mockVolumesClient) Create(ctx context.Context, projectID string, body types.BlockStorageRequest, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.BlockStorageResponse]{StatusCode: 200, Data: &types.BlockStorageResponse{}}, nil
}
func (m *mockVolumesClient) Update(ctx context.Context, projectID, id string, body types.BlockStorageRequest, params *types.RequestParameters) (*types.Response[types.BlockStorageResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, id, body, params)
	}
	return &types.Response[types.BlockStorageResponse]{StatusCode: 200, Data: &types.BlockStorageResponse{}}, nil
}
func (m *mockVolumesClient) Delete(ctx context.Context, projectID, id string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, id, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── project ──────────────────────────────────────────────────────────────────

type mockProjectClient struct {
	listFn   func(ctx context.Context, params *types.RequestParameters) (*types.Response[types.ProjectList], error)
	getFn    func(ctx context.Context, id string, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error)
	createFn func(ctx context.Context, body types.ProjectRequest, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error)
	updateFn func(ctx context.Context, id string, body types.ProjectRequest, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error)
	deleteFn func(ctx context.Context, id string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockProjectClient) List(ctx context.Context, params *types.RequestParameters) (*types.Response[types.ProjectList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, params)
	}
	return &types.Response[types.ProjectList]{StatusCode: 200, Data: &types.ProjectList{}}, nil
}
func (m *mockProjectClient) Get(ctx context.Context, id string, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, id, params)
	}
	return &types.Response[types.ProjectResponse]{StatusCode: 200, Data: &types.ProjectResponse{}}, nil
}
func (m *mockProjectClient) Create(ctx context.Context, body types.ProjectRequest, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, body, params)
	}
	return &types.Response[types.ProjectResponse]{StatusCode: 200, Data: &types.ProjectResponse{}}, nil
}
func (m *mockProjectClient) Update(ctx context.Context, id string, body types.ProjectRequest, params *types.RequestParameters) (*types.Response[types.ProjectResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, body, params)
	}
	return &types.Response[types.ProjectResponse]{StatusCode: 200, Data: &types.ProjectResponse{}}, nil
}
func (m *mockProjectClient) Delete(ctx context.Context, id string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── network: remaining sub-clients ──────────────────────────────────────────

type mockLoadBalancersClient struct {
	listFn func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerList], error)
	getFn  func(ctx context.Context, projectID string, lbID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error)
}

func (m *mockLoadBalancersClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.LoadBalancerList]{StatusCode: 200, Data: &types.LoadBalancerList{}}, nil
}
func (m *mockLoadBalancersClient) Get(ctx context.Context, projectID, lbID string, params *types.RequestParameters) (*types.Response[types.LoadBalancerResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, lbID, params)
	}
	return &types.Response[types.LoadBalancerResponse]{StatusCode: 200, Data: &types.LoadBalancerResponse{}}, nil
}

type mockSecurityGroupsClient struct {
	listFn   func(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[types.SecurityGroupList], error)
	getFn    func(ctx context.Context, projectID string, vpcID string, sgID string, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error)
	createFn func(ctx context.Context, projectID string, vpcID string, body types.SecurityGroupRequest, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error)
	updateFn func(ctx context.Context, projectID string, vpcID string, sgID string, body types.SecurityGroupRequest, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error)
	deleteFn func(ctx context.Context, projectID string, vpcID string, sgID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockSecurityGroupsClient) List(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[types.SecurityGroupList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, vpcID, params)
	}
	return &types.Response[types.SecurityGroupList]{StatusCode: 200, Data: &types.SecurityGroupList{}}, nil
}
func (m *mockSecurityGroupsClient) Get(ctx context.Context, projectID, vpcID, sgID string, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, vpcID, sgID, params)
	}
	return &types.Response[types.SecurityGroupResponse]{StatusCode: 200, Data: &types.SecurityGroupResponse{}}, nil
}
func (m *mockSecurityGroupsClient) Create(ctx context.Context, projectID, vpcID string, body types.SecurityGroupRequest, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, vpcID, body, params)
	}
	return &types.Response[types.SecurityGroupResponse]{StatusCode: 200, Data: &types.SecurityGroupResponse{}}, nil
}
func (m *mockSecurityGroupsClient) Update(ctx context.Context, projectID, vpcID, sgID string, body types.SecurityGroupRequest, params *types.RequestParameters) (*types.Response[types.SecurityGroupResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, vpcID, sgID, body, params)
	}
	return &types.Response[types.SecurityGroupResponse]{StatusCode: 200, Data: &types.SecurityGroupResponse{}}, nil
}
func (m *mockSecurityGroupsClient) Delete(ctx context.Context, projectID, vpcID, sgID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, vpcID, sgID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockSecurityGroupRulesClient struct {
	listFn   func(ctx context.Context, projectID string, vpcID string, sgID string, params *types.RequestParameters) (*types.Response[types.SecurityRuleList], error)
	getFn    func(ctx context.Context, projectID string, vpcID string, sgID string, ruleID string, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error)
	createFn func(ctx context.Context, projectID string, vpcID string, sgID string, body types.SecurityRuleRequest, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error)
	updateFn func(ctx context.Context, projectID string, vpcID string, sgID string, ruleID string, body types.SecurityRuleRequest, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error)
	deleteFn func(ctx context.Context, projectID string, vpcID string, sgID string, ruleID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockSecurityGroupRulesClient) List(ctx context.Context, projectID, vpcID, sgID string, params *types.RequestParameters) (*types.Response[types.SecurityRuleList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, vpcID, sgID, params)
	}
	return &types.Response[types.SecurityRuleList]{StatusCode: 200, Data: &types.SecurityRuleList{}}, nil
}
func (m *mockSecurityGroupRulesClient) Get(ctx context.Context, projectID, vpcID, sgID, ruleID string, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, vpcID, sgID, ruleID, params)
	}
	return &types.Response[types.SecurityRuleResponse]{StatusCode: 200, Data: &types.SecurityRuleResponse{}}, nil
}
func (m *mockSecurityGroupRulesClient) Create(ctx context.Context, projectID, vpcID, sgID string, body types.SecurityRuleRequest, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, vpcID, sgID, body, params)
	}
	return &types.Response[types.SecurityRuleResponse]{StatusCode: 200, Data: &types.SecurityRuleResponse{}}, nil
}
func (m *mockSecurityGroupRulesClient) Update(ctx context.Context, projectID, vpcID, sgID, ruleID string, body types.SecurityRuleRequest, params *types.RequestParameters) (*types.Response[types.SecurityRuleResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, vpcID, sgID, ruleID, body, params)
	}
	return &types.Response[types.SecurityRuleResponse]{StatusCode: 200, Data: &types.SecurityRuleResponse{}}, nil
}
func (m *mockSecurityGroupRulesClient) Delete(ctx context.Context, projectID, vpcID, sgID, ruleID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, vpcID, sgID, ruleID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockSubnetsClient struct {
	listFn   func(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[types.SubnetList], error)
	getFn    func(ctx context.Context, projectID string, vpcID string, subnetID string, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error)
	createFn func(ctx context.Context, projectID string, vpcID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error)
	updateFn func(ctx context.Context, projectID string, vpcID string, subnetID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error)
	deleteFn func(ctx context.Context, projectID string, vpcID string, subnetID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockSubnetsClient) List(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[types.SubnetList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, vpcID, params)
	}
	return &types.Response[types.SubnetList]{StatusCode: 200, Data: &types.SubnetList{}}, nil
}
func (m *mockSubnetsClient) Get(ctx context.Context, projectID, vpcID, subnetID string, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, vpcID, subnetID, params)
	}
	return &types.Response[types.SubnetResponse]{StatusCode: 200, Data: &types.SubnetResponse{}}, nil
}
func (m *mockSubnetsClient) Create(ctx context.Context, projectID, vpcID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, vpcID, body, params)
	}
	return &types.Response[types.SubnetResponse]{StatusCode: 200, Data: &types.SubnetResponse{}}, nil
}
func (m *mockSubnetsClient) Update(ctx context.Context, projectID, vpcID, subnetID string, body types.SubnetRequest, params *types.RequestParameters) (*types.Response[types.SubnetResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, vpcID, subnetID, body, params)
	}
	return &types.Response[types.SubnetResponse]{StatusCode: 200, Data: &types.SubnetResponse{}}, nil
}
func (m *mockSubnetsClient) Delete(ctx context.Context, projectID, vpcID, subnetID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, vpcID, subnetID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockVPCPeeringsClient struct {
	listFn   func(ctx context.Context, projectID string, vpcID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringList], error)
	getFn    func(ctx context.Context, projectID string, vpcID string, peeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error)
	createFn func(ctx context.Context, projectID string, vpcID string, body types.VPCPeeringRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error)
	updateFn func(ctx context.Context, projectID string, vpcID string, peeringID string, body types.VPCPeeringRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error)
	deleteFn func(ctx context.Context, projectID string, vpcID string, peeringID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockVPCPeeringsClient) List(ctx context.Context, projectID, vpcID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, vpcID, params)
	}
	return &types.Response[types.VPCPeeringList]{StatusCode: 200, Data: &types.VPCPeeringList{}}, nil
}
func (m *mockVPCPeeringsClient) Get(ctx context.Context, projectID, vpcID, peeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, vpcID, peeringID, params)
	}
	return &types.Response[types.VPCPeeringResponse]{StatusCode: 200, Data: &types.VPCPeeringResponse{}}, nil
}
func (m *mockVPCPeeringsClient) Create(ctx context.Context, projectID, vpcID string, body types.VPCPeeringRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, vpcID, body, params)
	}
	return &types.Response[types.VPCPeeringResponse]{StatusCode: 200, Data: &types.VPCPeeringResponse{}}, nil
}
func (m *mockVPCPeeringsClient) Update(ctx context.Context, projectID, vpcID, peeringID string, body types.VPCPeeringRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, vpcID, peeringID, body, params)
	}
	return &types.Response[types.VPCPeeringResponse]{StatusCode: 200, Data: &types.VPCPeeringResponse{}}, nil
}
func (m *mockVPCPeeringsClient) Delete(ctx context.Context, projectID, vpcID, peeringID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, vpcID, peeringID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockVPCPeeringRoutesClient struct {
	listFn   func(ctx context.Context, projectID string, vpcID string, peeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error)
	getFn    func(ctx context.Context, projectID string, vpcID string, peeringID string, routeID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error)
	createFn func(ctx context.Context, projectID string, vpcID string, peeringID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error)
	updateFn func(ctx context.Context, projectID string, vpcID string, peeringID string, routeID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error)
	deleteFn func(ctx context.Context, projectID string, vpcID string, peeringID string, routeID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockVPCPeeringRoutesClient) List(ctx context.Context, projectID, vpcID, peeringID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, vpcID, peeringID, params)
	}
	return &types.Response[types.VPCPeeringRouteList]{StatusCode: 200, Data: &types.VPCPeeringRouteList{}}, nil
}
func (m *mockVPCPeeringRoutesClient) Get(ctx context.Context, projectID, vpcID, peeringID, routeID string, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, vpcID, peeringID, routeID, params)
	}
	return &types.Response[types.VPCPeeringRouteResponse]{StatusCode: 200, Data: &types.VPCPeeringRouteResponse{}}, nil
}
func (m *mockVPCPeeringRoutesClient) Create(ctx context.Context, projectID, vpcID, peeringID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, vpcID, peeringID, body, params)
	}
	return &types.Response[types.VPCPeeringRouteResponse]{StatusCode: 200, Data: &types.VPCPeeringRouteResponse{}}, nil
}
func (m *mockVPCPeeringRoutesClient) Update(ctx context.Context, projectID, vpcID, peeringID, routeID string, body types.VPCPeeringRouteRequest, params *types.RequestParameters) (*types.Response[types.VPCPeeringRouteResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, vpcID, peeringID, routeID, body, params)
	}
	return &types.Response[types.VPCPeeringRouteResponse]{StatusCode: 200, Data: &types.VPCPeeringRouteResponse{}}, nil
}
func (m *mockVPCPeeringRoutesClient) Delete(ctx context.Context, projectID, vpcID, peeringID, routeID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, vpcID, peeringID, routeID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockVPNTunnelsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelList], error)
	getFn    func(ctx context.Context, projectID string, tunnelID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error)
	createFn func(ctx context.Context, projectID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error)
	updateFn func(ctx context.Context, projectID string, tunnelID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error)
	deleteFn func(ctx context.Context, projectID string, tunnelID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockVPNTunnelsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.VPNTunnelList]{StatusCode: 200, Data: &types.VPNTunnelList{}}, nil
}
func (m *mockVPNTunnelsClient) Get(ctx context.Context, projectID, tunnelID string, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, tunnelID, params)
	}
	return &types.Response[types.VPNTunnelResponse]{StatusCode: 200, Data: &types.VPNTunnelResponse{}}, nil
}
func (m *mockVPNTunnelsClient) Create(ctx context.Context, projectID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.VPNTunnelResponse]{StatusCode: 200, Data: &types.VPNTunnelResponse{}}, nil
}
func (m *mockVPNTunnelsClient) Update(ctx context.Context, projectID, tunnelID string, body types.VPNTunnelRequest, params *types.RequestParameters) (*types.Response[types.VPNTunnelResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, tunnelID, body, params)
	}
	return &types.Response[types.VPNTunnelResponse]{StatusCode: 200, Data: &types.VPNTunnelResponse{}}, nil
}
func (m *mockVPNTunnelsClient) Delete(ctx context.Context, projectID, tunnelID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, tunnelID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockVPNRoutesClient struct {
	listFn   func(ctx context.Context, projectID string, tunnelID string, params *types.RequestParameters) (*types.Response[types.VPNRouteList], error)
	getFn    func(ctx context.Context, projectID string, tunnelID string, routeID string, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error)
	createFn func(ctx context.Context, projectID string, tunnelID string, body types.VPNRouteRequest, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error)
	updateFn func(ctx context.Context, projectID string, tunnelID string, routeID string, body types.VPNRouteRequest, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error)
	deleteFn func(ctx context.Context, projectID string, tunnelID string, routeID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockVPNRoutesClient) List(ctx context.Context, projectID, tunnelID string, params *types.RequestParameters) (*types.Response[types.VPNRouteList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, tunnelID, params)
	}
	return &types.Response[types.VPNRouteList]{StatusCode: 200, Data: &types.VPNRouteList{}}, nil
}
func (m *mockVPNRoutesClient) Get(ctx context.Context, projectID, tunnelID, routeID string, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, tunnelID, routeID, params)
	}
	return &types.Response[types.VPNRouteResponse]{StatusCode: 200, Data: &types.VPNRouteResponse{}}, nil
}
func (m *mockVPNRoutesClient) Create(ctx context.Context, projectID, tunnelID string, body types.VPNRouteRequest, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, tunnelID, body, params)
	}
	return &types.Response[types.VPNRouteResponse]{StatusCode: 200, Data: &types.VPNRouteResponse{}}, nil
}
func (m *mockVPNRoutesClient) Update(ctx context.Context, projectID, tunnelID, routeID string, body types.VPNRouteRequest, params *types.RequestParameters) (*types.Response[types.VPNRouteResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, tunnelID, routeID, body, params)
	}
	return &types.Response[types.VPNRouteResponse]{StatusCode: 200, Data: &types.VPNRouteResponse{}}, nil
}
func (m *mockVPNRoutesClient) Delete(ctx context.Context, projectID, tunnelID, routeID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, tunnelID, routeID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── storage: remaining sub-clients ──────────────────────────────────────────

type mockSnapshotsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.SnapshotList], error)
	getFn    func(ctx context.Context, projectID string, snapshotID string, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	createFn func(ctx context.Context, projectID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	updateFn func(ctx context.Context, projectID string, snapshotID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error)
	deleteFn func(ctx context.Context, projectID string, snapshotID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockSnapshotsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.SnapshotList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.SnapshotList]{StatusCode: 200, Data: &types.SnapshotList{}}, nil
}
func (m *mockSnapshotsClient) Get(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, snapshotID, params)
	}
	return &types.Response[types.SnapshotResponse]{StatusCode: 200, Data: &types.SnapshotResponse{}}, nil
}
func (m *mockSnapshotsClient) Create(ctx context.Context, projectID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.SnapshotResponse]{StatusCode: 200, Data: &types.SnapshotResponse{}}, nil
}
func (m *mockSnapshotsClient) Update(ctx context.Context, projectID, snapshotID string, body types.SnapshotRequest, params *types.RequestParameters) (*types.Response[types.SnapshotResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, snapshotID, body, params)
	}
	return &types.Response[types.SnapshotResponse]{StatusCode: 200, Data: &types.SnapshotResponse{}}, nil
}
func (m *mockSnapshotsClient) Delete(ctx context.Context, projectID, snapshotID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, snapshotID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockStorageBackupsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.StorageBackupList], error)
	getFn    func(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)
	createFn func(ctx context.Context, projectID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)
	updateFn func(ctx context.Context, projectID string, backupID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error)
	deleteFn func(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockStorageBackupsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.StorageBackupList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.StorageBackupList]{StatusCode: 200, Data: &types.StorageBackupList{}}, nil
}
func (m *mockStorageBackupsClient) Get(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, backupID, params)
	}
	return &types.Response[types.StorageBackupResponse]{StatusCode: 200, Data: &types.StorageBackupResponse{}}, nil
}
func (m *mockStorageBackupsClient) Create(ctx context.Context, projectID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.StorageBackupResponse]{StatusCode: 200, Data: &types.StorageBackupResponse{}}, nil
}
func (m *mockStorageBackupsClient) Update(ctx context.Context, projectID, backupID string, body types.StorageBackupRequest, params *types.RequestParameters) (*types.Response[types.StorageBackupResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, backupID, body, params)
	}
	return &types.Response[types.StorageBackupResponse]{StatusCode: 200, Data: &types.StorageBackupResponse{}}, nil
}
func (m *mockStorageBackupsClient) Delete(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, backupID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockStorageRestoreClient struct {
	listFn   func(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[types.RestoreList], error)
	getFn    func(ctx context.Context, projectID string, backupID string, restoreID string, params *types.RequestParameters) (*types.Response[types.RestoreResponse], error)
	createFn func(ctx context.Context, projectID string, backupID string, body types.RestoreRequest, params *types.RequestParameters) (*types.Response[types.RestoreResponse], error)
	updateFn func(ctx context.Context, projectID string, backupID string, restoreID string, body types.RestoreRequest, params *types.RequestParameters) (*types.Response[types.RestoreResponse], error)
	deleteFn func(ctx context.Context, projectID string, backupID string, restoreID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockStorageRestoreClient) List(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.RestoreList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, backupID, params)
	}
	return &types.Response[types.RestoreList]{StatusCode: 200, Data: &types.RestoreList{}}, nil
}
func (m *mockStorageRestoreClient) Get(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, backupID, restoreID, params)
	}
	return &types.Response[types.RestoreResponse]{StatusCode: 200, Data: &types.RestoreResponse{}}, nil
}
func (m *mockStorageRestoreClient) Create(ctx context.Context, projectID, backupID string, body types.RestoreRequest, params *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, backupID, body, params)
	}
	return &types.Response[types.RestoreResponse]{StatusCode: 200, Data: &types.RestoreResponse{}}, nil
}
func (m *mockStorageRestoreClient) Update(ctx context.Context, projectID, backupID, restoreID string, body types.RestoreRequest, params *types.RequestParameters) (*types.Response[types.RestoreResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, backupID, restoreID, body, params)
	}
	return &types.Response[types.RestoreResponse]{StatusCode: 200, Data: &types.RestoreResponse{}}, nil
}
func (m *mockStorageRestoreClient) Delete(ctx context.Context, projectID, backupID, restoreID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, backupID, restoreID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── database ─────────────────────────────────────────────────────────────────

type mockDatabaseClient struct {
	dbaasClient     aruba.DBaaSClient
	databasesClient aruba.DatabasesClient
	backupsClient   aruba.BackupsClient
	usersClient     aruba.UsersClient
}

func (m *mockDatabaseClient) DBaaS() aruba.DBaaSClient         { return m.dbaasClient }
func (m *mockDatabaseClient) Databases() aruba.DatabasesClient { return m.databasesClient }
func (m *mockDatabaseClient) Backups() aruba.BackupsClient     { return m.backupsClient }
func (m *mockDatabaseClient) Users() aruba.UsersClient         { return m.usersClient }
func (m *mockDatabaseClient) Grants() aruba.GrantsClient       { return nil }

type mockDBaaSClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.DBaaSList], error)
	getFn    func(ctx context.Context, projectID string, dbaasID string, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error)
	createFn func(ctx context.Context, projectID string, body types.DBaaSRequest, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error)
	updateFn func(ctx context.Context, projectID string, dbaasID string, body types.DBaaSRequest, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error)
	deleteFn func(ctx context.Context, projectID string, dbaasID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockDBaaSClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.DBaaSList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.DBaaSList]{StatusCode: 200, Data: &types.DBaaSList{}}, nil
}
func (m *mockDBaaSClient) Get(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, dbaasID, params)
	}
	return &types.Response[types.DBaaSResponse]{StatusCode: 200, Data: &types.DBaaSResponse{}}, nil
}
func (m *mockDBaaSClient) Create(ctx context.Context, projectID string, body types.DBaaSRequest, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.DBaaSResponse]{StatusCode: 200, Data: &types.DBaaSResponse{}}, nil
}
func (m *mockDBaaSClient) Update(ctx context.Context, projectID, dbaasID string, body types.DBaaSRequest, params *types.RequestParameters) (*types.Response[types.DBaaSResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, dbaasID, body, params)
	}
	return &types.Response[types.DBaaSResponse]{StatusCode: 200, Data: &types.DBaaSResponse{}}, nil
}
func (m *mockDBaaSClient) Delete(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, dbaasID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockDatabasesClient struct {
	listFn   func(ctx context.Context, projectID string, dbaasID string, params *types.RequestParameters) (*types.Response[types.DatabaseList], error)
	getFn    func(ctx context.Context, projectID string, dbaasID string, dbID string, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error)
	createFn func(ctx context.Context, projectID string, dbaasID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error)
	updateFn func(ctx context.Context, projectID string, dbaasID string, dbID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error)
	deleteFn func(ctx context.Context, projectID string, dbaasID string, dbID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockDatabasesClient) List(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[types.DatabaseList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, dbaasID, params)
	}
	return &types.Response[types.DatabaseList]{StatusCode: 200, Data: &types.DatabaseList{}}, nil
}
func (m *mockDatabasesClient) Get(ctx context.Context, projectID, dbaasID, dbID string, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, dbaasID, dbID, params)
	}
	return &types.Response[types.DatabaseResponse]{StatusCode: 200, Data: &types.DatabaseResponse{}}, nil
}
func (m *mockDatabasesClient) Create(ctx context.Context, projectID, dbaasID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, dbaasID, body, params)
	}
	return &types.Response[types.DatabaseResponse]{StatusCode: 200, Data: &types.DatabaseResponse{}}, nil
}
func (m *mockDatabasesClient) Update(ctx context.Context, projectID, dbaasID, dbID string, body types.DatabaseRequest, params *types.RequestParameters) (*types.Response[types.DatabaseResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, dbaasID, dbID, body, params)
	}
	return &types.Response[types.DatabaseResponse]{StatusCode: 200, Data: &types.DatabaseResponse{}}, nil
}
func (m *mockDatabasesClient) Delete(ctx context.Context, projectID, dbaasID, dbID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, dbaasID, dbID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockDBBackupsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.BackupList], error)
	getFn    func(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[types.BackupResponse], error)
	createFn func(ctx context.Context, projectID string, body types.BackupRequest, params *types.RequestParameters) (*types.Response[types.BackupResponse], error)
	deleteFn func(ctx context.Context, projectID string, backupID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockDBBackupsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.BackupList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.BackupList]{StatusCode: 200, Data: &types.BackupList{}}, nil
}
func (m *mockDBBackupsClient) Get(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, backupID, params)
	}
	return &types.Response[types.BackupResponse]{StatusCode: 200, Data: &types.BackupResponse{}}, nil
}
func (m *mockDBBackupsClient) Create(ctx context.Context, projectID string, body types.BackupRequest, params *types.RequestParameters) (*types.Response[types.BackupResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.BackupResponse]{StatusCode: 200, Data: &types.BackupResponse{}}, nil
}
func (m *mockDBBackupsClient) Delete(ctx context.Context, projectID, backupID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, backupID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

type mockUsersClient struct {
	listFn   func(ctx context.Context, projectID string, dbaasID string, params *types.RequestParameters) (*types.Response[types.UserList], error)
	getFn    func(ctx context.Context, projectID string, dbaasID string, userID string, params *types.RequestParameters) (*types.Response[types.UserResponse], error)
	createFn func(ctx context.Context, projectID string, dbaasID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error)
	updateFn func(ctx context.Context, projectID string, dbaasID string, userID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error)
	deleteFn func(ctx context.Context, projectID string, dbaasID string, userID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockUsersClient) List(ctx context.Context, projectID, dbaasID string, params *types.RequestParameters) (*types.Response[types.UserList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, dbaasID, params)
	}
	return &types.Response[types.UserList]{StatusCode: 200, Data: &types.UserList{}}, nil
}
func (m *mockUsersClient) Get(ctx context.Context, projectID, dbaasID, userID string, params *types.RequestParameters) (*types.Response[types.UserResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, dbaasID, userID, params)
	}
	return &types.Response[types.UserResponse]{StatusCode: 200, Data: &types.UserResponse{}}, nil
}
func (m *mockUsersClient) Create(ctx context.Context, projectID, dbaasID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, dbaasID, body, params)
	}
	return &types.Response[types.UserResponse]{StatusCode: 200, Data: &types.UserResponse{}}, nil
}
func (m *mockUsersClient) Update(ctx context.Context, projectID, dbaasID, userID string, body types.UserRequest, params *types.RequestParameters) (*types.Response[types.UserResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, dbaasID, userID, body, params)
	}
	return &types.Response[types.UserResponse]{StatusCode: 200, Data: &types.UserResponse{}}, nil
}
func (m *mockUsersClient) Delete(ctx context.Context, projectID, dbaasID, userID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, dbaasID, userID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── container ────────────────────────────────────────────────────────────────

type mockContainerClient struct {
	kaasClient              aruba.KaaSClient
	containerRegistryClient aruba.ContainerRegistryClient
}

func (m *mockContainerClient) KaaS() aruba.KaaSClient { return m.kaasClient }
func (m *mockContainerClient) ContainerRegistry() aruba.ContainerRegistryClient {
	return m.containerRegistryClient
}

type mockKaaSClient struct {
	listFn               func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KaaSList], error)
	getFn                func(ctx context.Context, projectID string, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	createFn             func(ctx context.Context, projectID string, body types.KaaSRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	updateFn             func(ctx context.Context, projectID string, kaasID string, body types.KaaSUpdateRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error)
	deleteFn             func(ctx context.Context, projectID string, kaasID string, params *types.RequestParameters) (*types.Response[any], error)
	downloadKubeconfigFn func(ctx context.Context, projectID string, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error)
}

func (m *mockKaaSClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.KaaSList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.KaaSList]{StatusCode: 200, Data: &types.KaaSList{}}, nil
}
func (m *mockKaaSClient) Get(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, kaasID, params)
	}
	return &types.Response[types.KaaSResponse]{StatusCode: 200, Data: &types.KaaSResponse{}}, nil
}
func (m *mockKaaSClient) Create(ctx context.Context, projectID string, body types.KaaSRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.KaaSResponse]{StatusCode: 200, Data: &types.KaaSResponse{}}, nil
}
func (m *mockKaaSClient) Update(ctx context.Context, projectID, kaasID string, body types.KaaSUpdateRequest, params *types.RequestParameters) (*types.Response[types.KaaSResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, kaasID, body, params)
	}
	return &types.Response[types.KaaSResponse]{StatusCode: 200, Data: &types.KaaSResponse{}}, nil
}
func (m *mockKaaSClient) Delete(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, kaasID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}
func (m *mockKaaSClient) DownloadKubeconfig(ctx context.Context, projectID, kaasID string, params *types.RequestParameters) (*types.Response[types.KaaSKubeconfigResponse], error) {
	if m.downloadKubeconfigFn != nil {
		return m.downloadKubeconfigFn(ctx, projectID, kaasID, params)
	}
	return &types.Response[types.KaaSKubeconfigResponse]{StatusCode: 200, Data: &types.KaaSKubeconfigResponse{}}, nil
}

type mockContainerRegistryClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryList], error)
	getFn    func(ctx context.Context, projectID string, registryID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	createFn func(ctx context.Context, projectID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	updateFn func(ctx context.Context, projectID string, registryID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error)
	deleteFn func(ctx context.Context, projectID string, registryID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockContainerRegistryClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.ContainerRegistryList]{StatusCode: 200, Data: &types.ContainerRegistryList{}}, nil
}
func (m *mockContainerRegistryClient) Get(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, registryID, params)
	}
	return &types.Response[types.ContainerRegistryResponse]{StatusCode: 200, Data: &types.ContainerRegistryResponse{}}, nil
}
func (m *mockContainerRegistryClient) Create(ctx context.Context, projectID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.ContainerRegistryResponse]{StatusCode: 200, Data: &types.ContainerRegistryResponse{}}, nil
}
func (m *mockContainerRegistryClient) Update(ctx context.Context, projectID, registryID string, body types.ContainerRegistryRequest, params *types.RequestParameters) (*types.Response[types.ContainerRegistryResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, registryID, body, params)
	}
	return &types.Response[types.ContainerRegistryResponse]{StatusCode: 200, Data: &types.ContainerRegistryResponse{}}, nil
}
func (m *mockContainerRegistryClient) Delete(ctx context.Context, projectID, registryID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, registryID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── schedule ─────────────────────────────────────────────────────────────────

type mockScheduleClient struct {
	jobsClient aruba.JobsClient
}

func (m *mockScheduleClient) Jobs() aruba.JobsClient { return m.jobsClient }

type mockJobsClient struct {
	listFn   func(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.JobList], error)
	getFn    func(ctx context.Context, projectID string, jobID string, params *types.RequestParameters) (*types.Response[types.JobResponse], error)
	createFn func(ctx context.Context, projectID string, body types.JobRequest, params *types.RequestParameters) (*types.Response[types.JobResponse], error)
	updateFn func(ctx context.Context, projectID string, jobID string, body types.JobRequest, params *types.RequestParameters) (*types.Response[types.JobResponse], error)
	deleteFn func(ctx context.Context, projectID string, jobID string, params *types.RequestParameters) (*types.Response[any], error)
}

func (m *mockJobsClient) List(ctx context.Context, projectID string, params *types.RequestParameters) (*types.Response[types.JobList], error) {
	if m.listFn != nil {
		return m.listFn(ctx, projectID, params)
	}
	return &types.Response[types.JobList]{StatusCode: 200, Data: &types.JobList{}}, nil
}
func (m *mockJobsClient) Get(ctx context.Context, projectID, jobID string, params *types.RequestParameters) (*types.Response[types.JobResponse], error) {
	if m.getFn != nil {
		return m.getFn(ctx, projectID, jobID, params)
	}
	return &types.Response[types.JobResponse]{StatusCode: 200, Data: &types.JobResponse{}}, nil
}
func (m *mockJobsClient) Create(ctx context.Context, projectID string, body types.JobRequest, params *types.RequestParameters) (*types.Response[types.JobResponse], error) {
	if m.createFn != nil {
		return m.createFn(ctx, projectID, body, params)
	}
	return &types.Response[types.JobResponse]{StatusCode: 200, Data: &types.JobResponse{}}, nil
}
func (m *mockJobsClient) Update(ctx context.Context, projectID, jobID string, body types.JobRequest, params *types.RequestParameters) (*types.Response[types.JobResponse], error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, projectID, jobID, body, params)
	}
	return &types.Response[types.JobResponse]{StatusCode: 200, Data: &types.JobResponse{}}, nil
}
func (m *mockJobsClient) Delete(ctx context.Context, projectID, jobID string, params *types.RequestParameters) (*types.Response[any], error) {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, projectID, jobID, params)
	}
	return &types.Response[any]{StatusCode: 200}, nil
}

// ─── security (KMS is a concrete SDK type — cannot be mocked externally) ─────

type mockSecurityClient struct{}

// KMS returns nil. The KMS SDK client (aruba.KMSClient = *security.KMSClientWrapper)
// is a concrete internal type that cannot be constructed outside the SDK vendor tree.
// KMS command tests are therefore not supported via this mock infrastructure.
func (m *mockSecurityClient) KMS() aruba.KMSClient { return nil }

// ─── helpers ─────────────────────────────────────────────────────────────────

// strPtr returns a pointer to s. Convenience for building SDK response structs in tests.
func strPtr(s string) *string { return &s }

// newMockClient constructs a mockClient wired to the given sub-mocks.
// Pass nil for sub-mocks that the test doesn't exercise.
func newMockClient(opts ...func(*mockClient)) *mockClient {
	c := &mockClient{}
	for _, o := range opts {
		o(c)
	}
	return c
}

// withNetwork attaches a network client built with the supplied VPCs mock.
func withNetwork(vpcs *mockVPCsClient) func(*mockClient) {
	return func(c *mockClient) {
		c.networkClient = &mockNetworkClient{vpcsMock: vpcs}
	}
}

// withCompute attaches a compute client built with the supplied CloudServers mock.
func withCompute(cs *mockCloudServersClient) func(*mockClient) {
	return func(c *mockClient) {
		c.computeClient = &mockComputeClient{cloudServersClient: cs}
	}
}

// withStorage attaches a storage client built with the supplied Volumes mock.
func withStorage(v *mockVolumesClient) func(*mockClient) {
	return func(c *mockClient) {
		c.storageClient = &mockStorageClient{volumesMock: v}
	}
}

// withProject attaches a project client mock.
func withProject(p *mockProjectClient) func(*mockClient) {
	return func(c *mockClient) {
		c.projectClient = p
	}
}

// withNetworkMock attaches a pre-built network mock (all sub-clients configurable).
func withNetworkMock(n *mockNetworkClient) func(*mockClient) {
	return func(c *mockClient) { c.networkClient = n }
}

// withComputeMock attaches a pre-built compute mock.
func withComputeMock(cm *mockComputeClient) func(*mockClient) {
	return func(c *mockClient) { c.computeClient = cm }
}

// withStorageMock attaches a pre-built storage mock.
func withStorageMock(s *mockStorageClient) func(*mockClient) {
	return func(c *mockClient) { c.storageClient = s }
}

// withDatabase attaches a database client mock.
func withDatabase(d *mockDatabaseClient) func(*mockClient) {
	return func(c *mockClient) { c.databaseClient = d }
}

// withContainer attaches a container client mock.
func withContainer(con *mockContainerClient) func(*mockClient) {
	return func(c *mockClient) { c.containerClient = con }
}

// withSchedule attaches a schedule client mock.
func withSchedule(s *mockScheduleClient) func(*mockClient) {
	return func(c *mockClient) { c.scheduleClient = s }
}

// resetCmdFlags walks the entire command tree and marks every flag as "not
// changed". This is required because cobra/pflag tracks a "changed" bit per
// flag and the global rootCmd is shared across test cases: without this reset,
// a flag set in test N remains "seen" in test N+1, causing MarkFlagRequired
// validation to pass even when the flag is absent.
func resetCmdFlags(cmd *cobra.Command) {
	resetFlagSet := func(f *pflag.Flag) {
		f.Changed = false
		if f.DefValue != "" {
			_ = f.Value.Set(f.DefValue) //nolint:errcheck
		}
	}
	cmd.Flags().VisitAll(resetFlagSet)
	cmd.PersistentFlags().VisitAll(resetFlagSet)
	for _, sub := range cmd.Commands() {
		resetCmdFlags(sub)
	}
}

// runCmd sets the mock client, executes rootCmd with the given args, then
// resets state. It returns the error from rootCmd.Execute().
func runCmd(mock aruba.Client, args []string) error {
	resetCmdFlags(rootCmd)
	setClientForTesting(mock)
	defer resetClientState()
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// errSDK returns a non-nil SDK error suitable for injection in test cases.
func errSDK(msg string) error { return errors.New(msg) }

// runCmdCapture executes rootCmd with the given args (via runCmd) and returns
// everything written to os.Stdout during the call. Use this instead of runCmd
// when a test needs to assert output content.
func runCmdCapture(mock aruba.Client, args []string) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	runErr := runCmd(mock, args)
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r) //nolint:errcheck
	r.Close()
	return buf.String(), runErr
}
