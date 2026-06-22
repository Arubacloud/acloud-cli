package cmd

import (
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// makeProjectCmd returns a minimal cobra.Command with --project-id set.
func makeProjectCmd(projectID string) *cobra.Command {
	cmd := &cobra.Command{}
	_ = cmd.Flags().String("project-id", "", "")
	_ = cmd.Flags().Set("project-id", projectID)
	return cmd
}

// ─── completeCloudServerID ──────────────────────────────────────────────────

func TestCompleteCloudServerID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeCloudServerID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

// ─── completeKeyPairID ──────────────────────────────────────────────────────

func TestCompleteKeyPairID(t *testing.T) {
	name := "my-keypair"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/keyPairs", jsonResponse(200, types.KeyPairListResponse{
		Values: []types.KeyPairResponse{
			{Metadata: types.ResourceMetadataResponse{Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeKeyPairID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteKeyPairID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeKeyPairID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

// ─── completeContainerRegistryID ────────────────────────────────────────────

func TestCompleteContainerRegistryID(t *testing.T) {
	id, name := "cr-001", "my-registry"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryListResponse{
		Values: []types.ContainerRegistryResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeContainerRegistryID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteContainerRegistryID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeContainerRegistryID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeKaaSID ─────────────────────────────────────────────────────────

func TestCompleteKaaSID(t *testing.T) {
	id, name := "kaas-001", "my-cluster"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas", jsonResponse(200, types.KaaSListResponse{
		Values: []types.KaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeKaaSID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteKaaSID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeKaaSID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeDatabaseBackupID ───────────────────────────────────────────────

func TestCompleteDatabaseBackupID(t *testing.T) {
	id, name := "bkp-001", "my-backup"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/backups", jsonResponse(200, types.DBaaSBackupListResponse{
		Values: []types.BackupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDatabaseBackupID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteDatabaseBackupID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeDatabaseBackupID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeDBaaSDatabaseID ────────────────────────────────────────────────

func TestCompleteDBaaSDatabaseID_NoArgs(t *testing.T) {
	comps, _ := completeDBaaSDatabaseID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions with no args, got %v", comps)
	}
}

func TestCompleteDBaaSDatabaseID(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
		Values: []types.DatabaseResponse{
			{Name: "my-db"},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaaSDatabaseID(makeProjectCmd("proj-123"), []string{"dbaas-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeDBaaSID ────────────────────────────────────────────────────────

func TestCompleteDBaaSID(t *testing.T) {
	id, name := "dbaas-001", "my-dbaas"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaaSID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteDBaaSID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeDBaaSID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeGrantID ────────────────────────────────────────────────────────

func TestCompleteGrantID_NoArgs(t *testing.T) {
	comps, _ := completeGrantID(&cobra.Command{}, []string{"only-one"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with insufficient args, got %v", comps)
	}
}

func TestCompleteGrantID(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases/mydb/grants", jsonResponse(200, types.GrantListResponse{
		Values: []types.GrantResponse{
			{User: types.GrantUserCommon{Username: "alice"}, Role: types.GrantRoleCommon{Name: "admin"}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeGrantID(makeProjectCmd("proj-123"), []string{"dbaas-001", "mydb"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeDBaaSUserID ────────────────────────────────────────────────────

func TestCompleteDBaaSUserID_NoArgs(t *testing.T) {
	comps, _ := completeDBaaSUserID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions with no args, got %v", comps)
	}
}

func TestCompleteDBaaSUserID(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/users", jsonResponse(200, types.DatabaseUserListResponse{
		Values: []types.UserResponse{
			{Username: "testuser"},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaaSUserID(makeProjectCmd("proj-123"), []string{"dbaas-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeProjectID ──────────────────────────────────────────────────────

func TestCompleteProjectID(t *testing.T) {
	id, name := "proj-001", "my-project"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects", jsonResponse(200, types.ProjectListResponse{
		Values: []types.ProjectResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeProjectID(&cobra.Command{}, nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteProjectID_NoClient(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()
	resetClientState()

	comps, _ := completeProjectID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions without client, got %v", comps)
	}
}

// ─── completeElasticIPID ────────────────────────────────────────────────────

func TestCompleteElasticIPID(t *testing.T) {
	id, name := "eip-001", "my-eip"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps", jsonResponse(200, types.ElasticIPListResponse{
		Values: []types.ElasticIPResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeElasticIPID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteElasticIPID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeElasticIPID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeLoadBalancerID ─────────────────────────────────────────────────

func TestCompleteLoadBalancerID(t *testing.T) {
	id, name := "lb-001", "my-lb"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/loadBalancers", jsonResponse(200, types.LoadBalancerListResponse{
		Values: []types.LoadBalancerResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeLoadBalancerID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteLoadBalancerID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeLoadBalancerID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeSecurityRuleID ─────────────────────────────────────────────────

func TestCompleteSecurityRuleID_InsufficientArgs(t *testing.T) {
	comps, _ := completeSecurityRuleID(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with insufficient args, got %v", comps)
	}
}

func TestCompleteSecurityRuleID(t *testing.T) {
	id, name := "rule-001", "my-rule"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups/sg-001/securityRules", jsonResponse(200, types.SecurityRuleListResponse{
		Values: []types.SecurityRuleResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityRuleID(makeProjectCmd("proj-123"), []string{"vpc-001", "sg-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeVPCID ──────────────────────────────────────────────────────────

func TestCompleteVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteVPCID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeVPCID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeVPCPeeringRouteID ──────────────────────────────────────────────

func TestCompleteVPCPeeringRouteID_InsufficientArgs(t *testing.T) {
	comps, _ := completeVPCPeeringRouteID(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with insufficient args, got %v", comps)
	}
}

func TestCompleteVPCPeeringRouteID(t *testing.T) {
	id, name := "route-001", "my-route"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings/peer-001/vpcPeeringRoutes", jsonResponse(200, types.VPCPeeringRouteListResponse{
		Values: []types.VPCPeeringRouteResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringRouteID(makeProjectCmd("proj-123"), []string{"vpc-001", "peer-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeVPNRouteID ─────────────────────────────────────────────────────

func TestCompleteVPNRouteID_NoArgs(t *testing.T) {
	comps, _ := completeVPNRouteID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions with no args, got %v", comps)
	}
}

func TestCompleteVPNRouteID(t *testing.T) {
	id, name := "vroute-001", "my-route"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels/vpn-001/vpnRoutes", jsonResponse(200, types.VPNRouteListResponse{
		Values: []types.VPNRouteResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPNRouteID(makeProjectCmd("proj-123"), []string{"vpn-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeVPNTunnelID ────────────────────────────────────────────────────

func TestCompleteVPNTunnelID(t *testing.T) {
	id, name := "vpn-001", "my-tunnel"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelListResponse{
		Values: []types.VPNTunnelResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPNTunnelID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteVPNTunnelID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeVPNTunnelID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeJobID ──────────────────────────────────────────────────────────

func TestCompleteJobID(t *testing.T) {
	id, name := "job-001", "my-job"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Schedule/jobs", jsonResponse(200, types.JobListResponse{
		Values: []types.JobResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeJobID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteJobID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeJobID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeKMSID ──────────────────────────────────────────────────────────

func TestCompleteKMSID(t *testing.T) {
	id, name := "kms-001", "my-kms"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsListResponse{
		Values: []types.KmsResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeKMSID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteKMSID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeKMSID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeBackupID (storage) ─────────────────────────────────────────────

func TestCompleteBackupID(t *testing.T) {
	id, name := "bkp-001", "my-backup"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupListResponse{
		Values: []types.StorageBackupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeBackupID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteBackupID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeBackupID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeBlockStorageID ─────────────────────────────────────────────────

func TestCompleteBlockStorageID(t *testing.T) {
	id, name := "vol-001", "my-volume"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
		Values: []types.BlockStorageResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeBlockStorageID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteBlockStorageID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeBlockStorageID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── completeRestoreID ──────────────────────────────────────────────────────

func TestCompleteRestoreID_NoArgs_DelegatesToBackup(t *testing.T) {
	id, name := "bkp-001", "my-backup"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupListResponse{
		Values: []types.StorageBackupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	// When no args: delegates to completeBackupID
	completions, _ := completeRestoreID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions from backup delegation, got none")
	}
}

func TestCompleteRestoreID_TooManyArgs(t *testing.T) {
	comps, _ := completeRestoreID(&cobra.Command{}, []string{"bkp-001", "rst-001"}, "")
	if comps != nil {
		t.Errorf("expected nil with too many args, got %v", comps)
	}
}

func TestCompleteRestoreID(t *testing.T) {
	id, name := "rst-001", "my-restore"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups/bkp-001/restores", jsonResponse(200, types.StorageRestoreListResponse{
		Values: []types.StorageRestoreResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeRestoreID(makeProjectCmd("proj-123"), []string{"bkp-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

// ─── completeSnapshotID ─────────────────────────────────────────────────────

func TestCompleteSnapshotID(t *testing.T) {
	id, name := "snap-001", "my-snap"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/snapshots", jsonResponse(200, types.SnapshotListResponse{
		Values: []types.SnapshotResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSnapshotID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteSnapshotID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeSnapshotID(&cobra.Command{}, nil, "")
	if comps != nil {
		t.Errorf("expected nil completions, got %v", comps)
	}
}

// ─── prefix-filter coverage (#184) ──────────────────────────────────────────

// TestCompleteBlockStorageID_PrefixFilter verifies that completeBlockStorageID
// only returns IDs that start with the provided prefix.
func TestCompleteBlockStorageID_PrefixFilter(t *testing.T) {
	vol1, vol2 := "aaa-001", "bbb-002"
	name1, name2 := "vol-aaa", "vol-bbb"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
		Values: []types.BlockStorageResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &vol1, Name: &name1}},
			{Metadata: types.ResourceMetadataResponse{ID: &vol2, Name: &name2}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeBlockStorageID(makeProjectCmd("proj-123"), nil, "aaa")
	if len(completions) != 1 {
		t.Fatalf("prefix 'aaa': expected 1 completion, got %d: %v", len(completions), completions)
	}
	if completions[0][:7] != "aaa-001" {
		t.Errorf("expected completion starting with aaa-001, got %q", completions[0])
	}
}

// TestCompleteVPCID_PrefixFilter verifies prefix filtering in the network family.
func TestCompleteVPCID_PrefixFilter(t *testing.T) {
	vpc1, vpc2 := "vpc-alpha", "vpc-beta"
	n1, n2 := "alpha-vpc", "beta-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &vpc1, Name: &n1}},
			{Metadata: types.ResourceMetadataResponse{ID: &vpc2, Name: &n2}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCID(makeProjectCmd("proj-123"), nil, "vpc-b")
	if len(completions) != 1 {
		t.Fatalf("prefix 'vpc-b': expected 1 completion, got %d: %v", len(completions), completions)
	}
	if completions[0][:8] != "vpc-beta" {
		t.Errorf("expected completion starting with vpc-beta, got %q", completions[0])
	}
}

// TestCompleteCloudServerID_PrefixNoMatch verifies that an unmatched prefix
// returns an empty (not nil) slice.
func TestCompleteCloudServerID_PrefixNoMatch(t *testing.T) {
	csID, csName := "cs-001", "my-server"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Compute/cloudServers", jsonResponse(200, types.CloudServerListResponse{
		Values: []types.CloudServerResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &csID, Name: &csName}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeCloudServerID(makeProjectCmd("proj-123"), nil, "zzz")
	if len(completions) != 0 {
		t.Errorf("expected 0 completions for unmatched prefix, got %v", completions)
	}
}

// ─── API-error coverage (#184) ───────────────────────────────────────────────

// TestCompleteBlockStorageID_APIError verifies that an API error returns nil
// completions rather than panicking or leaking errors.
func TestCompleteBlockStorageID_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages",
		errorResponse(500, "Internal Server Error", "storage unavailable"))
	setClientForTesting(srv.Client())
	defer resetClientState()

	comps, _ := completeBlockStorageID(makeProjectCmd("proj-123"), nil, "")
	if comps != nil {
		t.Errorf("expected nil completions on API error, got %v", comps)
	}
}

// TestCompleteDBaaSID_APIError verifies the database family error path.
func TestCompleteDBaaSID_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas",
		errorResponse(503, "Service Unavailable", "db not ready"))
	setClientForTesting(srv.Client())
	defer resetClientState()

	comps, _ := completeDBaaSID(makeProjectCmd("proj-123"), nil, "")
	if comps != nil {
		t.Errorf("expected nil completions on API error, got %v", comps)
	}
}

// TestCompleteKaaSID_APIError verifies the container family error path.
func TestCompleteKaaSID_APIError(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/kaas",
		errorResponse(503, "Service Unavailable", "unavailable"))
	setClientForTesting(srv.Client())
	defer resetClientState()

	comps, _ := completeKaaSID(makeProjectCmd("proj-123"), nil, "")
	if comps != nil {
		t.Errorf("expected nil completions on API error, got %v", comps)
	}
}

// ─── completeSubnetID ────────────────────────────────────────────────────────

func TestCompleteSubnetID_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSubnetID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs from delegation, got none")
	}
}

func TestCompleteSubnetID_NoArgs_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeSubnetID(&cobra.Command{}, []string{}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteSubnetID(t *testing.T) {
	id, name := "sub-001", "my-subnet"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/subnets", jsonResponse(200, types.SubnetListResponse{
		Values: []types.SubnetResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSubnetID(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteSubnetID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeSubnetID(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteSubnetID_TooManyArgs(t *testing.T) {
	comps, _ := completeSubnetID(&cobra.Command{}, []string{"vpc-001", "sub-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeSecurityGroupID ─────────────────────────────────────────────────

func TestCompleteSecurityGroupID_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityGroupID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs from delegation, got none")
	}
}

func TestCompleteSecurityGroupID_NoArgs_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeSecurityGroupID(&cobra.Command{}, []string{}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteSecurityGroupID(t *testing.T) {
	id, name := "sg-001", "my-sg"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
		Values: []types.SecurityGroupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityGroupID(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteSecurityGroupID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeSecurityGroupID(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteSecurityGroupID_TooManyArgs(t *testing.T) {
	comps, _ := completeSecurityGroupID(&cobra.Command{}, []string{"vpc-001", "sg-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeVPCPeeringID ────────────────────────────────────────────────────

func TestCompleteVPCPeeringID_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs from delegation, got none")
	}
}

func TestCompleteVPCPeeringID_NoArgs_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeVPCPeeringID(&cobra.Command{}, []string{}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteVPCPeeringID(t *testing.T) {
	id, name := "peer-001", "my-peering"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings", jsonResponse(200, types.VPCPeeringListResponse{
		Values: []types.VPCPeeringResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringID(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected completions, got none")
	}
}

func TestCompleteVPCPeeringID_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeVPCPeeringID(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteVPCPeeringID_TooManyArgs(t *testing.T) {
	comps, _ := completeVPCPeeringID(&cobra.Command{}, []string{"vpc-001", "peer-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeSecurityRuleID parent-slot fixes ────────────────────────────────

func TestCompleteSecurityRuleID_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityRuleID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs from delegation, got none")
	}
}

func TestCompleteSecurityRuleID_OneArg_DelegatesToSGID(t *testing.T) {
	id, name := "sg-001", "my-sg"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
		Values: []types.SecurityGroupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityRuleID(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected SG IDs from delegation, got none")
	}
}

func TestCompleteSecurityRuleID_TooManyArgs(t *testing.T) {
	comps, _ := completeSecurityRuleID(&cobra.Command{}, []string{"vpc-001", "sg-001", "rule-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeVPCPeeringRouteID parent-slot fixes ─────────────────────────────

func TestCompleteVPCPeeringRouteID_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringRouteID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs from delegation, got none")
	}
}

func TestCompleteVPCPeeringRouteID_OneArg_DelegatesToPeeringID(t *testing.T) {
	id, name := "peer-001", "my-peering"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings", jsonResponse(200, types.VPCPeeringListResponse{
		Values: []types.VPCPeeringResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringRouteID(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected peering IDs from delegation, got none")
	}
}

func TestCompleteVPCPeeringRouteID_TooManyArgs(t *testing.T) {
	comps, _ := completeVPCPeeringRouteID(&cobra.Command{}, []string{"vpc-001", "peer-001", "route-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeDBaaSDatabaseID parent-slot fixes ───────────────────────────────

func TestCompleteDBaaSDatabaseID_NoArgs_DelegatesToDBaaSID(t *testing.T) {
	id, name := "dbaas-001", "my-dbaas"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaaSDatabaseID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected DBaaS IDs from delegation, got none")
	}
}

func TestCompleteDBaaSDatabaseID_NoArgs_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeDBaaSDatabaseID(&cobra.Command{}, []string{}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteDBaaSDatabaseID_TooManyArgs(t *testing.T) {
	comps, _ := completeDBaaSDatabaseID(&cobra.Command{}, []string{"dbaas-001", "my-db"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeDBaaSUserID parent-slot fixes ────────────────────────────────────

func TestCompleteDBaaSUserID_NoArgs_DelegatesToDBaaSID(t *testing.T) {
	id, name := "dbaas-001", "my-dbaas"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaaSUserID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected DBaaS IDs from delegation, got none")
	}
}

func TestCompleteDBaaSUserID_NoArgs_NoProjectID(t *testing.T) {
	cleanup := withTempHomeDir(t)
	defer cleanup()

	comps, _ := completeDBaaSUserID(&cobra.Command{}, []string{}, "")
	if comps != nil {
		t.Errorf("expected nil completions without project-id, got %v", comps)
	}
}

func TestCompleteDBaaSUserID_TooManyArgs(t *testing.T) {
	comps, _ := completeDBaaSUserID(&cobra.Command{}, []string{"dbaas-001", "alice"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeGrantID parent-slot fixes ───────────────────────────────────────

func TestCompleteGrantID_NoArgs_DelegatesToDBaaSID(t *testing.T) {
	id, name := "dbaas-001", "my-dbaas"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeGrantID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected DBaaS IDs from delegation, got none")
	}
}

func TestCompleteGrantID_OneArg_DelegatesToDatabaseID(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas/dbaas-001/databases", jsonResponse(200, types.DatabaseListResponse{
		Values: []types.DatabaseResponse{
			{Name: "my-db"},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeGrantID(makeProjectCmd("proj-123"), []string{"dbaas-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected database names from delegation, got none")
	}
}

func TestCompleteGrantID_TooManyArgs(t *testing.T) {
	comps, _ := completeGrantID(&cobra.Command{}, []string{"dbaas-001", "my-db", "grant-id"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeStorageRestoreCreateArgs ────────────────────────────────────────

func TestCompleteStorageRestoreCreateArgs_NoArgs_DelegatesToBackupID(t *testing.T) {
	id, name := "bkp-001", "my-backup"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/backups", jsonResponse(200, types.StorageBackupListResponse{
		Values: []types.StorageBackupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeStorageRestoreCreateArgs(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected backup IDs from delegation, got none")
	}
}

func TestCompleteStorageRestoreCreateArgs_OneArg_DelegatesToBlockStorageID(t *testing.T) {
	id, name := "vol-001", "my-volume"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
		Values: []types.BlockStorageResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeStorageRestoreCreateArgs(makeProjectCmd("proj-123"), []string{"bkp-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected volume IDs from delegation, got none")
	}
}

func TestCompleteStorageRestoreCreateArgs_TooManyArgs(t *testing.T) {
	comps, _ := completeStorageRestoreCreateArgs(&cobra.Command{}, []string{"bkp-001", "vol-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── storageBackupCmd create ValidArgsFunction ───────────────────────────────

func TestCompleteBackupCreateArg_DelegatesToBlockStorageID(t *testing.T) {
	id, name := "vol-001", "my-volume"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageListResponse{
		Values: []types.BlockStorageResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeBlockStorageID(makeProjectCmd("proj-123"), nil, "")
	if len(completions) == 0 {
		t.Errorf("expected volume IDs, got none")
	}
}

// ─── completeVPNRouteID fix: now delegates to VPN tunnel on first arg ─────────

func TestCompleteVPNRouteID_NoArgs_DelegatesToVPNTunnelID(t *testing.T) {
	id, name := "vpn-001", "my-tunnel"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelListResponse{
		Values: []types.VPNTunnelResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPNRouteID(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPN tunnel IDs from delegation, got none")
	}
}

func TestCompleteVPNRouteID_TooManyArgs(t *testing.T) {
	comps, _ := completeVPNRouteID(&cobra.Command{}, []string{"vpn-001", "route-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions with too many args, got %v", comps)
	}
}

// ─── completeSubnetCreate ────────────────────────────────────────────────────

func TestCompleteSubnetCreate_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSubnetCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs, got none")
	}
}

func TestCompleteSubnetCreate_WithArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeSubnetCreate(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after vpc-id is provided, got %v", comps)
	}
}

// ─── completeSecurityGroupCreate ─────────────────────────────────────────────

func TestCompleteSecurityGroupCreate_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityGroupCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs, got none")
	}
}

func TestCompleteSecurityGroupCreate_WithArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeSecurityGroupCreate(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after vpc-id is provided, got %v", comps)
	}
}

// ─── completeSecurityRuleCreate ──────────────────────────────────────────────

func TestCompleteSecurityRuleCreate_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityRuleCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs, got none")
	}
}

func TestCompleteSecurityRuleCreate_OneArg_DelegatesToSGID(t *testing.T) {
	id, name := "sg-001", "my-sg"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/securityGroups", jsonResponse(200, types.SecurityGroupListResponse{
		Values: []types.SecurityGroupResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeSecurityRuleCreate(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected SG IDs, got none")
	}
}

func TestCompleteSecurityRuleCreate_TwoArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeSecurityRuleCreate(&cobra.Command{}, []string{"vpc-001", "sg-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after sg-id is provided, got %v", comps)
	}
}

// ─── completeVPCPeeringCreate ────────────────────────────────────────────────

func TestCompleteVPCPeeringCreate_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs, got none")
	}
}

func TestCompleteVPCPeeringCreate_WithArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeVPCPeeringCreate(&cobra.Command{}, []string{"vpc-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after vpc-id is provided, got %v", comps)
	}
}

// ─── completeVPCPeeringRouteCreate ───────────────────────────────────────────

func TestCompleteVPCPeeringRouteCreate_NoArgs_DelegatesToVPCID(t *testing.T) {
	id, name := "vpc-001", "my-vpc"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCListResponse{
		Values: []types.VPCResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringRouteCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPC IDs, got none")
	}
}

func TestCompleteVPCPeeringRouteCreate_OneArg_DelegatesToPeeringID(t *testing.T) {
	id, name := "peer-001", "my-peering"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-001/vpcPeerings", jsonResponse(200, types.VPCPeeringListResponse{
		Values: []types.VPCPeeringResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPCPeeringRouteCreate(makeProjectCmd("proj-123"), []string{"vpc-001"}, "")
	if len(completions) == 0 {
		t.Errorf("expected peering IDs, got none")
	}
}

func TestCompleteVPCPeeringRouteCreate_TwoArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeVPCPeeringRouteCreate(&cobra.Command{}, []string{"vpc-001", "peer-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after peering-id is provided, got %v", comps)
	}
}

// ─── completeVPNRouteCreate ──────────────────────────────────────────────────

func TestCompleteVPNRouteCreate_NoArgs_DelegatesToVPNTunnelID(t *testing.T) {
	id, name := "vpn-001", "my-tunnel"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpnTunnels", jsonResponse(200, types.VPNTunnelListResponse{
		Values: []types.VPNTunnelResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeVPNRouteCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected VPN tunnel IDs, got none")
	}
}

func TestCompleteVPNRouteCreate_WithArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeVPNRouteCreate(&cobra.Command{}, []string{"vpn-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after vpn-tunnel-id is provided, got %v", comps)
	}
}

// ─── completeDBaasDatabaseCreate ─────────────────────────────────────────────

func TestCompleteDBaasDatabaseCreate_NoArgs_DelegatesToDBaaSID(t *testing.T) {
	id, name := "dbaas-001", "my-dbaas"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaasDatabaseCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected DBaaS IDs, got none")
	}
}

func TestCompleteDBaasDatabaseCreate_WithArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeDBaasDatabaseCreate(&cobra.Command{}, []string{"dbaas-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after dbaas-id is provided, got %v", comps)
	}
}

// ─── completeDBaaSUserCreate ─────────────────────────────────────────────────

func TestCompleteDBaaSUserCreate_NoArgs_DelegatesToDBaaSID(t *testing.T) {
	id, name := "dbaas-001", "my-dbaas"
	srv := newArubaTestServer(t)
	srv.OnGet("/projects/proj-123/providers/Aruba.Database/dbaas", jsonResponse(200, types.DBaaSListResponse{
		Values: []types.DBaaSResponse{
			{Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name}},
		},
	}))
	setClientForTesting(srv.Client())
	defer resetClientState()

	completions, _ := completeDBaaSUserCreate(makeProjectCmd("proj-123"), []string{}, "")
	if len(completions) == 0 {
		t.Errorf("expected DBaaS IDs, got none")
	}
}

func TestCompleteDBaaSUserCreate_WithArgs_ReturnsNil(t *testing.T) {
	comps, _ := completeDBaaSUserCreate(&cobra.Command{}, []string{"dbaas-001"}, "")
	if comps != nil {
		t.Errorf("expected nil completions after dbaas-id is provided, got %v", comps)
	}
}
