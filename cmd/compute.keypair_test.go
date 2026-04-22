package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/Arubacloud/sdk-go/pkg/types"
)

func TestKeyPairListCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKeyPairsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with results",
			setupMock: func(m *mockKeyPairsClient) {
				kpID, kpName := "kp-001", "my-kp"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
					return &types.Response[types.KeyPairListResponse]{
						StatusCode: 200,
						Data: &types.KeyPairListResponse{
							Values: []types.KeyPairResponse{
								{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "success empty",
			setupMock: func(m *mockKeyPairsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
					return &types.Response[types.KeyPairListResponse]{StatusCode: 200, Data: &types.KeyPairListResponse{}}, nil
				}
			},
		},
		{
			name: "skips entries with nil ID",
			setupMock: func(m *mockKeyPairsClient) {
				kpID, kpName := "kp-001", "my-kp"
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
					return &types.Response[types.KeyPairListResponse]{
						StatusCode: 200,
						Data: &types.KeyPairListResponse{
							Values: []types.KeyPairResponse{
								{Metadata: types.ResourceMetadataResponse{Name: &kpName}},
								{Metadata: types.ResourceMetadataResponse{ID: &kpID, Name: &kpName}},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKeyPairsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
					return nil, fmt.Errorf("connection refused")
				}
			},
			wantErr:     true,
			errContains: "listing",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKeyPairsClient) {
				m.listFn = func(_ context.Context, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairListResponse], error) {
					return &types.Response[types.KeyPairListResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockKeyPairsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withComputeMock(&mockComputeClient{keyPairsClient: m})),
				[]string{"compute", "keypair", "list", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKeyPairGetCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKeyPairsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			setupMock: func(m *mockKeyPairsClient) {
				kpName := "my-kp"
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
					return &types.Response[types.KeyPairResponse]{
						StatusCode: 200,
						Data:       &types.KeyPairResponse{Metadata: types.ResourceMetadataResponse{Name: &kpName}},
					}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKeyPairsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "getting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKeyPairsClient) {
				m.getFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
					return &types.Response[types.KeyPairResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockKeyPairsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withComputeMock(&mockComputeClient{keyPairsClient: m})),
				[]string{"compute", "keypair", "get", "kp-001", "--project-id", "proj-123"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKeyPairCreateCmd(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		setupMock   func(*mockKeyPairsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			setupMock: func(m *mockKeyPairsClient) {
				kpName := "my-kp"
				m.createFn = func(_ context.Context, _ string, _ types.KeyPairRequest, _ *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
					return &types.Response[types.KeyPairResponse]{
						StatusCode: 200,
						Data:       &types.KeyPairResponse{Metadata: types.ResourceMetadataResponse{Name: &kpName}},
					}, nil
				}
			},
		},
		{
			name:        "missing required flag --name",
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--public-key", "ssh-rsa AAAA"},
			wantErr:     true,
			errContains: "name",
		},
		{
			name:        "missing required flag --public-key",
			args:        []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp"},
			wantErr:     true,
			errContains: "public-key",
		},
		{
			name: "SDK error propagates",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			setupMock: func(m *mockKeyPairsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.KeyPairRequest, _ *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
					return nil, fmt.Errorf("duplicate name")
				}
			},
			wantErr:     true,
			errContains: "creating",
		},
		{
			name: "API error propagates",
			args: []string{"compute", "keypair", "create", "--project-id", "proj-123", "--name", "my-kp", "--public-key", "ssh-rsa AAAA"},
			setupMock: func(m *mockKeyPairsClient) {
				m.createFn = func(_ context.Context, _ string, _ types.KeyPairRequest, _ *types.RequestParameters) (*types.Response[types.KeyPairResponse], error) {
					return &types.Response[types.KeyPairResponse]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockKeyPairsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withComputeMock(&mockComputeClient{keyPairsClient: m})), tc.args)
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}

func TestKeyPairDeleteCmd(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockKeyPairsClient)
		wantErr     bool
		errContains string
	}{
		{
			name: "success with --yes",
			setupMock: func(m *mockKeyPairsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{StatusCode: 200}, nil
				}
			},
		},
		{
			name: "SDK error propagates",
			setupMock: func(m *mockKeyPairsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantErr:     true,
			errContains: "deleting",
		},
		{
			name: "API error propagates",
			setupMock: func(m *mockKeyPairsClient) {
				m.deleteFn = func(_ context.Context, _, _ string, _ *types.RequestParameters) (*types.Response[any], error) {
					return &types.Response[any]{
						StatusCode: 404,
						Error:      &types.ErrorResponse{Title: strPtr("Not Found"), Detail: strPtr("resource not found")},
					}, nil
				}
			},
			wantErr:     true,
			errContains: "API error (status 404): Not Found",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockKeyPairsClient{}
			if tc.setupMock != nil {
				tc.setupMock(m)
			}
			err := runCmd(newMockClient(withComputeMock(&mockComputeClient{keyPairsClient: m})),
				[]string{"compute", "keypair", "delete", "kp-001", "--project-id", "proj-123", "--yes"})
			checkErr(t, err, tc.wantErr, tc.errContains)
		})
	}
}
