package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func securityRuleListPayload(l *aruba.List[*aruba.SecurityRule]) any {
	if r, ok := l.Raw().(*types.Response[types.SecurityRuleList]); ok && r != nil {
		return r.Data
	}
	return nil
}

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
	securityruleCreateCmd.MarkFlagRequired("name")
	securityruleCreateCmd.MarkFlagRequired("region")
	securityruleCreateCmd.MarkFlagRequired("direction")
	securityruleCreateCmd.MarkFlagRequired("protocol")
	securityruleCreateCmd.MarkFlagRequired("target-kind")
	securityruleCreateCmd.MarkFlagRequired("target-value")

	securityruleGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	securityruleUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleUpdateCmd.Flags().String("name", "", "New name for the security rule")
	securityruleUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	securityruleDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	securityruleDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	securityruleListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securityruleListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	securityruleListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	securityruleGetCmd.ValidArgsFunction = completeSecurityRuleID
	securityruleUpdateCmd.ValidArgsFunction = completeSecurityRuleID
	securityruleDeleteCmd.ValidArgsFunction = completeSecurityRuleID
}

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
	list, err := client.FromNetwork().SecurityGroupRules().List(ctx, securityGroupRef(projectID, vpcID, securityGroupID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, rule := range list.Items() {
			id := rule.SecurityRuleID()
			name := rule.Name()
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

// SecurityRule subcommands
var securityruleCmd = &cobra.Command{
	Use:   "securityrule",
	Short: "Manage security rules",
	Long:  `Perform CRUD operations on security rules in Aruba Cloud.`,
}

var securityruleCreateCmd = &cobra.Command{
	Use:   "create [vpc-id] [securitygroup-id]",
	Short: "Create a new security rule",
	Long: `Create a new security rule in the specified security group.

Valid values:
  --direction   Ingress (inbound) or Egress (outbound)
  --protocol    ANY, TCP, UDP, or ICMP
  --target-kind Ip (CIDR block) or SecurityGroup (another security group ID)
  --port        Port number or range (e.g., "80", "8000-8080"); omit for ANY/ICMP

Example target values:
  --target-kind Ip --target-value 10.0.0.0/8
  --target-kind SecurityGroup --target-value <sg-id>`,
	Example: `  # Allow inbound HTTP from any IP
  acloud network securityrule create <vpc-id> <sg-id> \
    --name allow-http --region IT-BG \
    --direction Ingress --protocol TCP --port 80 \
    --target-kind Ip --target-value 0.0.0.0/0

  # Allow all outbound traffic
  acloud network securityrule create <vpc-id> <sg-id> \
    --name allow-all-out --region IT-BG \
    --direction Egress --protocol ANY \
    --target-kind Ip --target-value 0.0.0.0/0`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

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

		rule := aruba.NewSecurityRule().
			IntoSecurityGroup(securityGroupRef(projectID, vpcID, securityGroupID)).
			Named(name).
			InRegion(aruba.Region(region)).
			WithDirection(types.RuleDirection(direction)).
			WithProtocol(aruba.RuleProtocol(protocol)).
			WithPort(port)
		if targetKind == string(types.EndpointTypeSecurityGroup) {
			rule.WithTargetSecurityGroup(aruba.URI(targetValue))
		} else {
			rule.WithTargetCIDR(targetValue)
		}
		if len(tags) > 0 {
			rule.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromNetwork().SecurityGroupRules().Create(ctx, rule)
		if err != nil {
			return fmt.Errorf("creating security rule: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil && r.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "DIRECTION", Width: 12},
				{Header: "PROTOCOL", Width: 12},
				{Header: "PORT", Width: 12},
				{Header: "STATUS", Width: 15},
			}
			status := ""
			if r.Status.State != nil {
				status = *r.Status.State
			}
			PrintOutput(r, headers, [][]string{{name, *r.Metadata.ID, direction, protocol, port, status}})
		} else {
			fmt.Println(msgCreatedAsync("Security rule", name))
		}
		return nil
	},
}

var securityruleGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Get security rule details",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		securityGroupID := args[1]
		securityRuleID := args[2]

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
		got, err := client.FromNetwork().SecurityGroupRules().Get(ctx, aruba.SecurityRuleRef(projectID, vpcID, securityGroupID, securityRuleID))
		if err != nil {
			return fmt.Errorf("getting security rule: %w", apiErrFromV2(err))
		}

		rule := got.Raw()
		if rule != nil {
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
				fmt.Printf("Creation Date:   %s\n", rule.Metadata.CreationDate.Format(DateLayout))
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
		return nil
	},
}

var securityruleListCmd = &cobra.Command{
	Use:   "list [vpc-id] [securitygroup-id]",
	Short: "List security rules for a security group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		securityGroupID := args[1]

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
		list, err := client.FromNetwork().SecurityGroupRules().List(ctx, securityGroupRef(projectID, vpcID, securityGroupID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing security rules: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
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
			for _, rule := range list.Items() {
				r := rule.Raw()
				name := rule.Name()
				id := rule.SecurityRuleID()
				direction := ""
				protocol := ""
				port := ""
				target := ""
				status := ""
				if r != nil {
					direction = string(r.Properties.Direction)
					protocol = string(r.Properties.Protocol)
					port = r.Properties.Port
					if r.Properties.Target != nil {
						target = fmt.Sprintf("%s:%s", r.Properties.Target.Kind, r.Properties.Target.Value)
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}
				rows = append(rows, []string{name, id, direction, protocol, port, target, status})
			}
			PrintOutput(securityRuleListPayload(list), headers, rows)
		} else {
			fmt.Println("No security rules found.")
		}
		return nil
	},
}

var securityruleUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Update a security rule",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		securityGroupID := args[1]
		securityRuleID := args[2]

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one field (--name or --tags) must be provided for update")
		}

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

		cur, err := client.FromNetwork().SecurityGroupRules().Get(ctx, aruba.SecurityRuleRef(projectID, vpcID, securityGroupID, securityRuleID))
		if err != nil {
			return fmt.Errorf("fetching current security rule: %w", apiErrFromV2(err))
		}
		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("security rule not found")
		}
		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update security rule while it is in 'InCreation' state. Please wait until the security rule is fully created")
		}

		if cur.Region() == "" {
			vpc, verr := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(projectID, vpcID))
			if verr == nil && vpc != nil && vpc.Raw() != nil && vpc.Raw().Metadata.LocationResponse != nil {
				cur.InRegion(vpc.Raw().Metadata.LocationResponse.Value)
			}
		}
		if cur.Region() == "" {
			return fmt.Errorf("unable to determine region value for security rule. Please ensure the VPC has a valid region")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}

		debugEnabled, _ := rootCmd.PersistentFlags().GetBool("debug")
		if debugEnabled {
			fmt.Fprintf(os.Stderr, "\n=== DEBUG: Security Rule Update Request ===\n")
			fmt.Fprintf(os.Stderr, "VPC ID: %s\n", vpcID)
			fmt.Fprintf(os.Stderr, "Security Group ID: %s\n", securityGroupID)
			fmt.Fprintf(os.Stderr, "Security Rule ID: %s\n", securityRuleID)
			fmt.Fprintf(os.Stderr, "Request Payload:\n")
			if reqJSON, err := json.MarshalIndent(cur.RawRequest(), "", "  "); err == nil {
				fmt.Fprintf(os.Stderr, "%s\n", reqJSON)
			}
			fmt.Fprintf(os.Stderr, "==========================================\n\n")
		}

		updated, err := client.FromNetwork().SecurityGroupRules().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating security rule: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "DIRECTION", Width: 12},
				{Header: "PROTOCOL", Width: 12},
				{Header: "PORT", Width: 12},
				{Header: "STATUS", Width: 15},
			}
			updName := ""
			if ur.Metadata.Name != nil {
				updName = *ur.Metadata.Name
			}
			updID := ""
			if ur.Metadata.ID != nil {
				updID = *ur.Metadata.ID
			}
			updStatus := ""
			if ur.Status.State != nil {
				updStatus = *ur.Status.State
			}
			PrintOutput(ur, headers, [][]string{{updName, updID, string(ur.Properties.Direction), string(ur.Properties.Protocol), ur.Properties.Port, updStatus}})
		} else {
			fmt.Println(msgUpdatedAsync("Security rule", securityRuleID))
		}
		return nil
	},
}

var securityruleDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [securitygroup-id] [securityrule-id]",
	Short: "Delete a security rule",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		securityGroupID := args[1]
		securityRuleID := args[2]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("security rule", securityRuleID)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

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

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			_, err = client.FromNetwork().SecurityGroupRules().Get(ctx, aruba.SecurityRuleRef(projectID, vpcID, securityGroupID, securityRuleID))
			if err != nil {
				return fmt.Errorf("dry-run: security rule not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("security rule", securityRuleID))
			return nil
		}

		if err := client.FromNetwork().SecurityGroupRules().Delete(ctx, aruba.SecurityRuleRef(projectID, vpcID, securityGroupID, securityRuleID)); err != nil {
			return fmt.Errorf("deleting security rule: %w", apiErrFromV2(err))
		}

		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		result := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{securityRuleID, status}
		PrintOutput(result, headers, [][]string{{securityRuleID, status}})
		return nil
	},
}
