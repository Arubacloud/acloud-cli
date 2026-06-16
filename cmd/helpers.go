package cmd

import (
	"context"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

// projectRef returns an opaque aruba.Ref for /projects/<projectID>.
func projectRef(projectID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID)
}

// projectWrapper fetches the *aruba.Project for projectID so callers can chain
// InProject(proj) on project-scoped resources.
func projectWrapper(ctx context.Context, client aruba.Client, projectID string) (*aruba.Project, error) {
	proj, err := client.FromProject().Get(ctx, projectRef(projectID))
	if err != nil {
		return nil, apiErrFromV2(err)
	}
	return proj, nil
}

// listOpts builds pagination CallOptions from --limit and --offset flags.
// Returns nil when neither flag is set, preserving the nil-means-no-options contract.
func listOpts(cmd *cobra.Command) []aruba.CallOption {
	limit, _ := cmd.Flags().GetInt32("limit")
	offset, _ := cmd.Flags().GetInt32("offset")
	if limit == 0 && offset == 0 {
		return nil
	}
	var opts []aruba.CallOption
	if limit > 0 {
		opts = append(opts, aruba.WithLimit(int(limit)))
	}
	if offset > 0 {
		opts = append(opts, aruba.WithOffset(int(offset)))
	}
	return opts
}

// extractIDFromURI returns the last path segment of a URI string.
// e.g. "/projects/p-1/providers/Aruba.Network/elasticIps/eip-42" → "eip-42"
func extractIDFromURI(uri string) string {
	uri = strings.TrimRight(uri, "/")
	idx := strings.LastIndex(uri, "/")
	if idx < 0 {
		return uri
	}
	return uri[idx+1:]
}
