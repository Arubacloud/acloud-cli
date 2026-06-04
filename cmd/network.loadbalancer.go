package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
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
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/loadBalancers/" + lbID)
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
	list, err := client.FromNetwork().LoadBalancers().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, lb := range list.Items() {
			id := lb.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, lb.Name()))
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
	RunE:  runLoadBalancerList,
}

var loadbalancerGetCmd = &cobra.Command{
	Use:   "get <loadbalancer-id>",
	Short: "Get Load Balancer details",
	Args:  cobra.ExactArgs(1),
	RunE:  runLoadBalancerGet,
}

func runLoadBalancerList(cmd *cobra.Command, args []string) error {
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
	list, err := client.FromNetwork().LoadBalancers().List(ctx, aruba.URI("/projects/"+projectID))
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
			raw := lb.Raw()
			if raw == nil {
				continue
			}
			name := ""
			if raw.Metadata.Name != nil {
				name = *raw.Metadata.Name
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			region := ""
			if raw.Metadata.LocationResponse != nil {
				region = string(raw.Metadata.LocationResponse.Value)
			}
			address := ""
			if raw.Properties.Address != nil {
				address = *raw.Properties.Address
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, region, address, status})
		}

		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No Load Balancers found")
	}
	return nil
}

func runLoadBalancerGet(cmd *cobra.Command, args []string) error {
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
	lb, err := client.FromNetwork().LoadBalancers().Get(ctx, aruba.LoadBalancerRef(projectID, lbID))
	if err != nil {
		return fmt.Errorf("getting Load Balancer details: %w", apiErrFromV2(err))
	}

	if lb != nil && lb.Raw() != nil {
		raw := lb.Raw()

		fmt.Println("\nLoad Balancer Details:")
		fmt.Println("======================")

		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Properties.Address != nil {
			fmt.Printf("Address:         %s\n", *raw.Properties.Address)
		}
		if raw.Properties.VPC != nil && raw.Properties.VPC.URI != "" {
			fmt.Printf("VPC:             %s\n", raw.Properties.VPC.URI)
		}

		fmt.Printf("Linked Resources: %d\n", len(raw.Properties.LinkedResources))

		if raw.Metadata.CreationDate != nil {
			fmt.Printf("Creation Date:   %s\n", raw.Metadata.CreationDate.Format(DateLayout))
		}
		if raw.Metadata.CreatedBy != nil {
			fmt.Printf("Created By:      %s\n", *raw.Metadata.CreatedBy)
		}

		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
		} else {
			fmt.Printf("Tags:            []\n")
		}

		if raw.Status.State != nil {
			fmt.Printf("Status:          %s\n", *raw.Status.State)
		}
	}
	return nil
}
