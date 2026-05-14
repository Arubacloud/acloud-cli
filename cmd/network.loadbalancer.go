package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	// LoadBalancer (read-only)
	networkCmd.AddCommand(loadbalancerCmd)
	loadbalancerCmd.AddCommand(loadbalancerListCmd)
	loadbalancerCmd.AddCommand(loadbalancerGetCmd)

	// Add flags for Load Balancer commands
	loadbalancerGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	loadbalancerListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	loadbalancerListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	loadbalancerListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	loadbalancerGetCmd.ValidArgsFunction = completeLoadBalancerID

}

func loadBalancerRef(projectID, lbID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/loadbalancers/" + lbID)
}

func loadBalancerListPayload(l *aruba.List[*aruba.LoadBalancer]) any {
	if r, ok := l.Raw().(*types.Response[types.LoadBalancerList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// Completion functions for network resources
func completeLoadBalancerID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().LoadBalancers().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, lb := range list.Items() {
			id := lb.LoadBalancerID()
			name := lb.Name()
			if id == "" {
				continue
			}
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, name))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// LoadBalancer subcommands
var loadbalancerCmd = &cobra.Command{
	Use:   "loadbalancer",
	Short: "Manage Load Balancers",
	Long:  `View Load Balancers in Aruba Cloud. Load Balancers are read-only resources.`,
}

var loadbalancerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Load Balancers",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		list, err := client.FromNetwork().LoadBalancers().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing Load Balancers: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "ADDRESS", Width: 16},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, lb := range list.Items() {
				r := lb.Raw()
				name := lb.Name()
				id := lb.LoadBalancerID()

				region := ""
				address := ""
				status := ""
				if r != nil {
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					if r.Properties.Address != nil {
						address = *r.Properties.Address
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}

				rows = append(rows, []string{name, id, region, address, status})
			}

			PrintOutput(loadBalancerListPayload(list), headers, rows)
		} else {
			fmt.Println("No Load Balancers found")
		}
		return nil
	},
}

var loadbalancerGetCmd = &cobra.Command{
	Use:   "get <loadbalancer-id>",
	Short: "Get Load Balancer details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lbID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		got, err := client.FromNetwork().LoadBalancers().Get(ctx, loadBalancerRef(projectID, lbID))
		if err != nil {
			return fmt.Errorf("getting Load Balancer details: %w", apiErrFromV2(err))
		}

		lb := got.Raw()
		if lb != nil {
			fmt.Println("\nLoad Balancer Details:")
			fmt.Println("======================")

			if lb.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *lb.Metadata.ID)
			}
			if lb.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *lb.Metadata.URI)
			}
			if lb.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *lb.Metadata.Name)
			}
			if lb.Properties.Address != nil {
				fmt.Printf("Address:         %s\n", *lb.Properties.Address)
			}
			if lb.Properties.VPC != nil && lb.Properties.VPC.URI != "" {
				fmt.Printf("VPC:             %s\n", lb.Properties.VPC.URI)
			}

			fmt.Printf("Linked Resources: %d\n", len(lb.Properties.LinkedResources))

			if lb.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", lb.Metadata.CreationDate.Format(DateLayout))
			}
			if lb.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *lb.Metadata.CreatedBy)
			}

			if len(lb.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", lb.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			if lb.Status.State != nil {
				fmt.Printf("Status:          %s\n", *lb.Status.State)
			}
		}
		return nil
	},
}
