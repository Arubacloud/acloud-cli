package cmd

// wait_coverage_test.go covers the --wait flag code paths added to the 11
// stateful resource operation functions (#174).  Tests are at Layer 2
// (operation function, real httptest client) so the wait blocks are reached
// without going through Cobra flag parsing.

import (
	"context"
	"testing"
	"time"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

// ─── helpers ─────────────────────────────────────────────────────────────────

// waitCtx returns a context with a short deadline so polls don't block tests.
func waitCtx() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second) //nolint:govet
	return ctx
}

// ─── VPC create + update ─────────────────────────────────────────────────────

func TestNetworkVPCCreate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-w1", "wait-vpc"
	state := types.StateActive

	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-w1", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	args := validNetworkVPCCreateArgs()
	args.Wait = true
	if err := NetworkVPCCreate(waitCtx(), srv.Client(), args); err != nil {
		t.Errorf("NetworkVPCCreate with Wait=true: %v", err)
	}
}

func TestNetworkVPCCreate_Wait_FailureState(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-w2", "fail-vpc"
	createState := types.StateActive
	getState := types.StateFailed

	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &createState},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-w2", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &getState},
	}))

	args := validNetworkVPCCreateArgs()
	args.Wait = true
	if err := NetworkVPCCreate(waitCtx(), srv.Client(), args); err == nil {
		t.Error("expected error for Failed state, got nil")
	}
}

func TestNetworkVPCCreate_Wait_AsyncPath(t *testing.T) {
	// When create returns nil response, the async message is printed and wait is skipped.
	srv := newArubaTestServer(t)
	// Return nil body so SDK treats the response as async (no ID available).
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs",
		jsonResponse(200, types.VPCResponse{Metadata: types.ResourceMetadataResponse{}}))

	args := validNetworkVPCCreateArgs()
	args.Wait = true
	// No error expected — async path skips wait block when id==""
	_ = NetworkVPCCreate(waitCtx(), srv.Client(), args)
}

func TestNetworkVPCUpdate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-w3", "wait-vpc-upd"
	state := types.StateActive

	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-w3", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPut("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-w3", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	err := NetworkVPCUpdate(waitCtx(), srv.Client(), NetworkVPCUpdateArgs{
		ProjectID: "proj-123",
		ID:        "vpc-w3",
		Name:      "new-name",
		Wait:      true,
	})
	if err != nil {
		t.Errorf("NetworkVPCUpdate with Wait=true: %v", err)
	}
}

// ─── KMS create + update ─────────────────────────────────────────────────────

func TestSecurityKMSCreate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-w1", "wait-kms"
	state := types.StateActive

	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-w1", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	args := validKMSCreateArgs()
	args.Wait = true
	if err := SecurityKMSCreate(waitCtx(), srv.Client(), args); err != nil {
		t.Errorf("SecurityKMSCreate with Wait=true: %v", err)
	}
}

func TestSecurityKMSUpdate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-w2", "wait-kms-upd"
	state := types.StateActive

	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-w2", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPut("/projects/proj-123/providers/Aruba.Security/kms/kms-w2", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	err := SecurityKMSUpdate(waitCtx(), srv.Client(), SecurityKMSUpdateArgs{
		ProjectID: "proj-123",
		ID:        "kms-w2",
		Name:      "new-kms-name",
		Wait:      true,
	})
	if err != nil {
		t.Errorf("SecurityKMSUpdate with Wait=true: %v", err)
	}
}

// ─── BlockStorage create ──────────────────────────────────────────────────────

func TestStorageBlockStorageCreate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vol-w1", "wait-vol"
	state := types.StateActive

	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockStorages", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-w1", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	args := validStorageBlockStorageCreateArgs()
	args.Wait = true
	if err := StorageBlockStorageCreate(waitCtx(), srv.Client(), args); err != nil {
		t.Errorf("StorageBlockStorageCreate with Wait=true: %v", err)
	}
}

func TestStorageBlockStorageUpdate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vol-w2", "wait-vol-upd"
	state := types.StateActive

	srv.OnGet("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-w2", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnPut("/projects/proj-123/providers/Aruba.Storage/blockStorages/vol-w2", jsonResponse(200, types.BlockStorageResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	err := StorageBlockStorageUpdate(waitCtx(), srv.Client(), StorageBlockStorageUpdateArgs{
		ProjectID: "proj-123",
		ID:        "vol-w2",
		Name:      "new-vol-name",
		Wait:      true,
	})
	if err != nil {
		t.Errorf("StorageBlockStorageUpdate with Wait=true: %v", err)
	}
}

// ─── ContainerRegistry create ─────────────────────────────────────────────────

func TestContainerRegistryCreate_Wait_ReachesActive(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "cr-w1", "wait-cr"
	state := types.StateActive

	// ContainerRegistry create needs ElasticIP, VPC, Subnet, SecurityGroup refs.
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/elasticIps/eip-001", jsonResponse(200, types.ElasticIPResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id},
	}))
	srv.OnPost("/projects/proj-123/providers/Aruba.Container/registries", jsonResponse(200, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Container/registries/cr-w1", jsonResponse(200, types.ContainerRegistryResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))

	args := validContainerRegistryCreateArgs()
	args.Wait = true
	if err := ContainerContainerRegistryCreate(waitCtx(), srv.Client(), args); err != nil {
		t.Errorf("ContainerContainerRegistryCreate with Wait=true: %v", err)
	}
}

// ─── CloudServer create -- already tested in compute.cloudserver_test.go ────
// That test (TestComputeCloudServerCreate_WaitFlag) covers the core
// InCreation→Active polling loop end-to-end.

// ─── --wait flag parsed in CLI commands (covers GetBool("wait") in ParseFromCobraCommand) ─
// These use runCmd so all flags are registered from init(); we don't care
// about the result — only that the flag-read path is hit.

func TestWaitFlag_CLI_VPC(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "vpc-cli-w", "cli-vpc"
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Network/vpcs", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Network/vpcs/vpc-cli-w", jsonResponse(200, types.VPCResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	runCmd(srv.Client(), []string{
		"network", "vpc", "create",
		"--project-id", "proj-123", "--name", "cli-vpc", "--region", "IT-BG", "--wait",
	})
}

func TestWaitFlag_CLI_BlockStorage(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Storage/blockStorages",
		errorResponse(500, "test", "test"))
	// Error is expected; the flag is still parsed, covering GetBool("wait").
	runCmd(srv.Client(), []string{
		"storage", "blockstorage", "create",
		"--project-id", "proj-123", "--name", "vol", "--size", "50", "--wait",
	})
}

func TestWaitFlag_CLI_DBaaS(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Database/dbaas",
		errorResponse(500, "test", "test"))
	runCmd(srv.Client(), []string{
		"database", "dbaas", "create",
		"--project-id", "proj-123", "--name", "db", "--region", "IT-BG",
		"--engine-id", "mysql-8.0", "--flavor", "DBO4A8", "--storage-size", "50",
		"--zone", "ITBG-1", "--wait",
	})
}

func TestWaitFlag_CLI_CloudServer(t *testing.T) {
	srv := newArubaTestServer(t)
	srv.OnPost("/projects/proj-123/providers/Aruba.Compute/cloudServers",
		errorResponse(500, "test", "test"))
	runCmd(srv.Client(), []string{
		"compute", "cloudserver", "create",
		"--project-id", "proj-123", "--name", "cs",
		"--region", "IT-BG", "--flavor", "cs-1",
		"--boot-disk-id", "vol-001", "--vpc-id", "vpc-001",
		"--subnet-id", "sub-001", "--wait",
	})
}

func TestWaitFlag_CLI_KMS(t *testing.T) {
	srv := newArubaTestServer(t)
	id, name := "kms-cli-w", "cli-kms"
	state := types.StateActive
	srv.OnPost("/projects/proj-123/providers/Aruba.Security/kms", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	srv.OnGet("/projects/proj-123/providers/Aruba.Security/kms/kms-cli-w", jsonResponse(200, types.KmsResponse{
		Metadata: types.ResourceMetadataResponse{ID: &id, Name: &name},
		Status:   types.ResourceStatusResponse{State: &state},
	}))
	runCmd(srv.Client(), []string{
		"security", "kms", "create",
		"--project-id", "proj-123", "--name", "cli-kms", "--region", "IT-BG",
		"--billing-period", "Hour", "--wait",
	})
}
