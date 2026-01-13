package cmd

import (
	"context"
	"fmt"
	"strings"

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

	// Set up auto-completion for resource IDs
	loadbalancerGetCmd.ValidArgsFunction = completeLoadBalancerID

}

// Completion functions for network resources
func completeLoadBalancerID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Allow completion even if args exist - user might be completing a partial ID

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	response, err := client.FromNetwork().LoadBalancers().List(ctx, projectID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, lb := range response.Data.Values {
			if lb.Metadata.ID != nil && lb.Metadata.Name != nil {
				id := *lb.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *lb.Metadata.Name))
				}
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
	Run: func(cmd *cobra.Command, args []string) {
		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// List Load Balancers using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().LoadBalancers().List(ctx, projectID, nil)
		if err != nil {
			fmt.Printf("Error listing Load Balancers: %v\n", err)
			return
		}

		if response != nil && response.Data != nil && len(response.Data.Values) > 0 {
			// Define table columns
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "ADDRESS", Width: 16},
				{Header: "STATUS", Width: 15},
			}

			// Build rows
			var rows [][]string
			for _, lb := range response.Data.Values {
				name := ""
				if lb.Metadata.Name != nil && *lb.Metadata.Name != "" {
					name = *lb.Metadata.Name
				}

				id := ""
				if lb.Metadata.ID != nil {
					id = *lb.Metadata.ID
				}

				region := lb.Metadata.LocationResponse.Value

				address := ""
				if lb.Properties.Address != nil {
					address = *lb.Properties.Address
				}

				status := ""
				if lb.Status.State != nil {
					status = *lb.Status.State
				}

				rows = append(rows, []string{name, id, region, address, status})
			}

			// Print the table
			PrintTable(headers, rows)
		} else {
			fmt.Println("No Load Balancers found")
		}
	},
}

var loadbalancerGetCmd = &cobra.Command{
	Use:   "get <loadbalancer-id>",
	Short: "Get Load Balancer details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		lbID := args[0]

		// Get project ID from flag or context
		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get SDK client
		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Get Load Balancer details using the SDK
		ctx := context.Background()
		response, err := client.FromNetwork().LoadBalancers().Get(ctx, projectID, lbID, nil)
		if err != nil {
			fmt.Printf("Error getting Load Balancer details: %v\n", err)
			return
		}

		if response != nil && response.Data != nil {
			lb := response.Data

			// Display Load Balancer details
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
				fmt.Printf("Creation Date:   %s\n", lb.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}
