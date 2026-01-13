package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {

	// SecurityRule
	networkCmd.AddCommand(securityruleCmd)
	securityruleCmd.AddCommand(securityruleCreateCmd)
	securityruleCmd.AddCommand(securityruleGetCmd)
	securityruleCmd.AddCommand(securityruleUpdateCmd)
	securityruleCmd.AddCommand(securityruleDeleteCmd)
	securityruleCmd.AddCommand(securityruleListCmd)

	// SecurityRule flags
	securityruleCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleCreateCmd.Flags().String("name", "", "Security Rule Name (required)")
	securityruleCreateCmd.Flags().String("region", "", "Region code (required)")
	securityruleCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	securityruleCreateCmd.Flags().String("direction", "", "Direction: Ingress or Egress (required)")
	securityruleCreateCmd.Flags().String("protocol", "", "Protocol: ANY, TCP, UDP, ICMP (required)")
	securityruleCreateCmd.Flags().String("port", "", "Port: a single numeric port, a port range or * (required for TCP/UDP)")
	securityruleCreateCmd.Flags().String("target-kind", "", "Target Kind: Ip or SecurityGroup (required)")
	securityruleCreateCmd.Flags().String("target-value", "", "Target Value: If kind = Ip, the value must be a valid network address in CIDR notation (included 0.0.0.0/0). If kind = SecurityGroup, the value must be a valid URI of any security group within the same VPC (required)")
	securityruleCreateCmd.Flags().BoolP("verbose", "v", false, "Show detailed debug information")

	securityruleGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	securityruleUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleUpdateCmd.Flags().String("name", "", "New name for the security rule")
	securityruleUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	securityruleDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")

	securityruleListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	// Set up auto-completion for resource IDs
	securityruleGetCmd.ValidArgsFunction = completeSecurityRuleID
	securityruleUpdateCmd.ValidArgsFunction = completeSecurityRuleID
	securityruleDeleteCmd.ValidArgsFunction = completeSecurityRuleID
}

// Completion functions for network resources
func completeSecurityRuleID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) < 2 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	vpcID := args[0]
	securityGroupID := args[1]

	ctx := context.Background()
	response, err := client.FromNetwork().SecurityGroupRules().List(ctx, projectID, vpcID, securityGroupID, nil)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if response != nil && response.Data != nil {
		for _, rule := range response.Data.Values {
			if rule.Metadata.ID != nil && rule.Metadata.Name != nil {
				id := *rule.Metadata.ID
				// Filter by partial input - use HasPrefix for more reliable matching
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *rule.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// SecurityRule subcommands
var securityruleCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Manage security rules",
	Long:  `Perform CRUD operations on security rules in Aruba Cloud.`,
}

var securityruleCreateCmd = &cobra.Command{
	Use:   "create [vpc-id] [securitygroup-id]",
	Short: "Create a new security rule",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		securityGroupID := args[1]

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		direction, _ := cmd.Flags().GetString("direction")
		protocol, _ := cmd.Flags().GetString("protocol")
		port, _ := cmd.Flags().GetString("port")
		targetKind, _ := cmd.Flags().GetString("target-kind")
		targetValue, _ := cmd.Flags().GetString("target-value")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Validate required fields
		if name == "" {
			fmt.Println("Error: --name is required")
			return
		}
		if region == "" {
			fmt.Println("Error: --region is required")
			return
		}
		if direction == "" {
			fmt.Println("Error: --direction is required")
			return
		}
		if protocol == "" {
			fmt.Println("Error: --protocol is required")
			return
		}
		if targetKind == "" {
			fmt.Println("Error: --target-kind is required")
			return
		}
		if targetValue == "" {
			fmt.Println("Error: --target-value is required")
			return
		}

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		// Build target
		target := &types.RuleTarget{
			Kind:  types.EndpointTypeDto(targetKind),
			Value: targetValue,
		}

		// Build the create request
		req := types.SecurityRuleRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: name,
					Tags: tags,
				},
				Location: types.LocationRequest{
					Value: region,
				},
			},
			Properties: types.SecurityRulePropertiesRequest{
				Direction: types.RuleDirection(direction),
				Protocol:  protocol,
				Port:      port,
				Target:    target,
			},
		}

		// Debug output if verbose
		if verbose {
			fmt.Println("Creating security rule with the following parameters:")
			fmt.Printf("  Name:         %s\n", name)
			fmt.Printf("  Region:       %s\n", region)
			fmt.Printf("  Direction:    %s\n", direction)
			fmt.Printf("  Protocol:     %s\n", protocol)
			fmt.Printf("  Port:         %s\n", port)
			fmt.Printf("  Target Kind:  %s\n", targetKind)
			fmt.Printf("  Target Value: %s\n", targetValue)
			if len(tags) > 0 {
				fmt.Printf("  Tags:         %v\n", tags)
			}
			fmt.Println()
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().SecurityGroupRules().Create(ctx, projectID, vpcID, securityGroupID, req, nil)
		if err != nil {
			fmt.Printf("Error creating security rule: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to create security rule - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil && resp.Data.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "DIRECTION", Width: 12},
				{Header: "PROTOCOL", Width: 12},
				{Header: "PORT", Width: 12},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				name,
				*resp.Data.Metadata.ID,
				direction,
				protocol,
				port,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Security rule created, but no ID returned.")
		}
	},
}

var securityruleGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Get security rule details",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		securityGroupID := args[1]
		securityRuleID := args[2]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, securityRuleID, nil)
		if err != nil {
			fmt.Printf("Error getting security rule: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get security rule - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil {
			rule := resp.Data
			fmt.Println("\nSecurity Rule Details:")
			fmt.Println("=====================")
			if rule.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *rule.Metadata.ID)
			}
			if rule.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *rule.Metadata.URI)
			}
			if rule.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *rule.Metadata.Name)
			}
			if rule.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", rule.Metadata.LocationResponse.Value)
			}
			fmt.Printf("Direction:       %s\n", rule.Properties.Direction)
			fmt.Printf("Protocol:        %s\n", rule.Properties.Protocol)
			fmt.Printf("Port:            %s\n", rule.Properties.Port)
			if rule.Properties.Target != nil {
				fmt.Printf("Target Kind:     %s\n", rule.Properties.Target.Kind)
				fmt.Printf("Target Value:    %s\n", rule.Properties.Target.Value)
			}
			if rule.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", rule.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
			}
			if rule.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *rule.Metadata.CreatedBy)
			}
			if len(rule.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", rule.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			if rule.Status.State != nil {
				fmt.Printf("Status:          %s\n", *rule.Status.State)
			}
		} else {
			fmt.Println("Security rule not found or no data returned.")
		}
	},
}

var securityruleListCmd = &cobra.Command{
	Use:   "list [vpc-id] [securitygroup-id]",
	Short: "List security rules for a security group",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		securityGroupID := args[1]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().SecurityGroupRules().List(ctx, projectID, vpcID, securityGroupID, nil)
		if err != nil {
			fmt.Printf("Error listing security rules: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list security rules - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "DIRECTION", Width: 12},
				{Header: "PROTOCOL", Width: 12},
				{Header: "PORT", Width: 12},
				{Header: "TARGET", Width: 30},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, rule := range resp.Data.Values {
				name := ""
				if rule.Metadata.Name != nil {
					name = *rule.Metadata.Name
				}
				id := ""
				if rule.Metadata.ID != nil {
					id = *rule.Metadata.ID
				}
				direction := string(rule.Properties.Direction)
				protocol := rule.Properties.Protocol
				port := rule.Properties.Port
				target := ""
				if rule.Properties.Target != nil {
					target = fmt.Sprintf("%s:%s", rule.Properties.Target.Kind, rule.Properties.Target.Value)
				}
				status := ""
				if rule.Status.State != nil {
					status = *rule.Status.State
				}
				rows = append(rows, []string{name, id, direction, protocol, port, target, status})
			}
			PrintTable(headers, rows)
		} else {
			fmt.Println("No security rules found.")
		}
	},
}

var securityruleUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Update a security rule",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		securityGroupID := args[1]
		securityRuleID := args[2]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		// At least one field must be provided
		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one field (--name or --tags) must be provided for update")
			return
		}

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()

		// Fetch current security rule details
		getResp, err := client.FromNetwork().SecurityGroupRules().Get(ctx, projectID, vpcID, securityGroupID, securityRuleID, nil)
		if err != nil {
			fmt.Printf("Error fetching current security rule: %v\n", err)
			return
		}

		if getResp == nil {
			fmt.Println("Error: No response received when fetching security rule")
			return
		}

		if getResp.IsError() && getResp.Error != nil {
			fmt.Printf("Failed to fetch security rule - Status: %d\n", getResp.StatusCode)
			if getResp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *getResp.Error.Title)
			}
			if getResp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *getResp.Error.Detail)
			}
			return
		}

		if getResp.Data == nil {
			fmt.Println("Error: Security rule not found or no data returned")
			return
		}

		current := getResp.Data

		// Block update if security rule is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update security rule while it is in 'InCreation' state. Please wait until the security rule is fully created.")
			return
		}

		// Get region value from current rule, or fetch from VPC if not available
		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}

		// If region value is not available from the rule, try to get it from the VPC
		if regionValue == "" {
			vpcResp, err := client.FromNetwork().VPCs().Get(ctx, projectID, vpcID, nil)
			if err == nil && vpcResp != nil && !vpcResp.IsError() && vpcResp.Data != nil {
				if vpcResp.Data.Metadata.LocationResponse != nil {
					regionValue = vpcResp.Data.Metadata.LocationResponse.Value
				}
			}
		}

		// If still no region value, we cannot proceed
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for security rule. Please ensure the VPC has a valid region.")
			return
		}

		// Build update request - only name and tags can be updated
		// Properties (direction, protocol, port, target) must remain unchanged
		req := types.SecurityRuleRequest{
			Metadata: types.RegionalResourceMetadataRequest{
				ResourceMetadataRequest: types.ResourceMetadataRequest{
					Name: func() string {
						if name != "" {
							return name
						}
						if current.Metadata.Name != nil {
							return *current.Metadata.Name
						}
						return ""
					}(),
					Tags: func() []string {
						if cmd.Flags().Changed("tags") {
							return tags
						}
						if current.Metadata.Tags != nil {
							return current.Metadata.Tags
						}
						return []string{}
					}(),
				},
				Location: types.LocationRequest{
					Value: regionValue,
				},
			},
			Properties: types.SecurityRulePropertiesRequest{
				// Properties cannot be updated - use current values
				Direction: current.Properties.Direction,
				Protocol:  current.Properties.Protocol,
				Port:      current.Properties.Port,
				Target:    current.Properties.Target,
			},
		}

		// Check if debug flag is enabled
		debugEnabled, _ := rootCmd.PersistentFlags().GetBool("debug")
		if debugEnabled {
			fmt.Fprintf(os.Stderr, "\n=== DEBUG: Security Rule Update Request ===\n")
			fmt.Fprintf(os.Stderr, "VPC ID: %s\n", vpcID)
			fmt.Fprintf(os.Stderr, "Security Group ID: %s\n", securityGroupID)
			fmt.Fprintf(os.Stderr, "Security Rule ID: %s\n", securityRuleID)
			fmt.Fprintf(os.Stderr, "Request Payload:\n")
			if reqJSON, err := json.MarshalIndent(req, "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", reqJSON)
			}
			fmt.Fprintf(os.Stderr, "==========================================\n\n")
		}

		resp, err := client.FromNetwork().SecurityGroupRules().Update(ctx, projectID, vpcID, securityGroupID, securityRuleID, req, nil)
		if err != nil {
			fmt.Printf("Error updating security rule: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to update security rule - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			if debugEnabled {
				fmt.Fprintf(os.Stderr, "\n=== DEBUG: Error Response ===\n")
				if resp.RawBody != nil {
					fmt.Fprintf(os.Stderr, "Raw Response Body:\n%s\n", string(resp.RawBody))
				}
				fmt.Fprintf(os.Stderr, "Status Code: %d\n", resp.StatusCode)
				fmt.Fprintf(os.Stderr, "===========================\n\n")
			}
			return
		}

		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "DIRECTION", Width: 12},
				{Header: "PROTOCOL", Width: 12},
				{Header: "PORT", Width: 12},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				func() string {
					if resp.Data.Metadata.Name != nil {
						return *resp.Data.Metadata.Name
					}
					return ""
				}(),
				func() string {
					if resp.Data.Metadata.ID != nil {
						return *resp.Data.Metadata.ID
					}
					return ""
				}(),
				string(resp.Data.Properties.Direction),
				resp.Data.Properties.Protocol,
				resp.Data.Properties.Port,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Printf("Security rule '%s' updated.\n", securityRuleID)
		}
	},
}

var securityruleDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Delete a security rule",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		securityGroupID := args[1]
		securityRuleID := args[2]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete security rule %s? This action cannot be undone.\n", securityRuleID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
		}

		client, err := GetArubaClient()
		if err != nil {
			fmt.Printf("Error initializing client: %v\n", err)
			return
		}

		ctx := context.Background()
		resp, err := client.FromNetwork().SecurityGroupRules().Delete(ctx, projectID, vpcID, securityGroupID, securityRuleID, nil)
		if err != nil {
			fmt.Printf("Error deleting security rule: %v\n", err)
			return
		}

		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to delete security rule - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}

		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		PrintTable(headers, [][]string{{securityRuleID, status}})
	},
}
