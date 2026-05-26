package cmd

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// securityGroupRef is shared with securityrule.go; uses hyphenated path segment (/security-groups/) to match SDK ID parser.
func securityGroupRef(projectID, vpcID, sgID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/vpcs/" + vpcID + "/security-groups/" + sgID)
}

func securityGroupListPayload(l *aruba.List[*aruba.SecurityGroup]) any {
	if r, ok := l.Raw().(*types.Response[types.SecurityGroupList]); ok && r != nil {
		return r.Data
	}
	return nil
}

func init() {
	// SecurityGroup
	networkCmd.AddCommand(securitygroupCmd)
	securitygroupCmd.AddCommand(securitygroupCreateCmd)
	securitygroupCmd.AddCommand(securitygroupGetCmd)
	securitygroupCmd.AddCommand(securitygroupUpdateCmd)
	securitygroupCmd.AddCommand(securitygroupDeleteCmd)
	securitygroupCmd.AddCommand(securitygroupListCmd)

	// SecurityGroup flags
	securitygroupCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupCreateCmd.Flags().String("name", "", "Security group name (required)")
	securitygroupCreateCmd.Flags().String("region", "", "Region code (required)")
	securitygroupCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	securitygroupCreateCmd.MarkFlagRequired("name")
	securitygroupCreateCmd.MarkFlagRequired("region")
	securitygroupGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupUpdateCmd.Flags().String("name", "", "New name for the security group")
	securitygroupUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	securitygroupDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	securitygroupDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	securitygroupListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	securitygroupListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	securitygroupListCmd.Flags().Int32("offset", 0, "Number of results to skip")
}

// SecurityGroup subcommands
var securitygroupCmd = &cobra.Command{
	Use:   "securitygroup",
	Short: "Manage security groups",
	Long:  `Perform CRUD operations on security groups in Aruba Cloud.`,
}

var securitygroupCreateCmd = &cobra.Command{
	Use:   "create [vpc-id]",
	Short: "Create a new security group",
	Long: `Create a new security group inside the specified VPC.

Security groups act as virtual firewalls. Add rules with 'acloud network securityrule create'
after the group is created.`,
	Example: `  acloud network securitygroup create <vpc-id> --name web-sg --region IT-BG
  acloud network securitygroup create <vpc-id> --name db-sg --region IT-BG --tags tier=db`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
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

		sg := aruba.NewSecurityGroup().
			InVPC(aruba.VPCRef(projectID, vpcID)).
			Named(name).
			RetaggedAs(tags...)

		resp, err := client.FromNetwork().SecurityGroups().Create(ctx, sg)
		if err != nil {
			return fmt.Errorf("creating security group: %w", apiErrFromV2(err))
		}
		if resp != nil && resp.Raw() != nil {
			raw := resp.Raw()
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			region := ""
			if raw.Metadata.LocationResponse != nil {
				region = string(raw.Metadata.LocationResponse.Value)
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			row := []string{name, id, region, status}
			PrintOutput(raw, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Security group", name))
		}
		return nil
	},
}

var securitygroupGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [securitygroup-id]",
	Short: "Get security group details",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		sgID := args[1]
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
		sg, err := client.FromNetwork().SecurityGroups().Get(ctx, aruba.SecurityGroupRef(projectID, vpcID, sgID))
		if err != nil {
			return fmt.Errorf("getting security group: %w", apiErrFromV2(err))
		}
		if sg != nil && sg.Raw() != nil {
			raw := sg.Raw()
			fmt.Println("\nSecurity Group Details:")
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
			if raw.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", string(raw.Metadata.LocationResponse.Value))
			}
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
		} else {
			fmt.Println("Security group not found or no data returned.")
		}
		return nil
	},
}

var securitygroupListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List security groups for a VPC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
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
		list, err := client.FromNetwork().SecurityGroups().List(ctx, aruba.VPCRef(projectID, vpcID))
		if err != nil {
			return fmt.Errorf("listing security groups: %w", apiErrFromV2(err))
		}
		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, sg := range list.Items() {
				raw := sg.Raw()
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
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, region, status})
			}
			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No security groups found.")
		}
		return nil
	},
}

var securitygroupUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [securitygroup-id]",
	Short: "Update a security group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		sgID := args[1]
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
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

		sg, err := client.FromNetwork().SecurityGroups().Get(ctx, aruba.SecurityGroupRef(projectID, vpcID, sgID))
		if err != nil || sg == nil || sg.Raw() == nil {
			return fmt.Errorf("fetching current security group: %w", apiErrFromV2(err))
		}

		if sg.Raw().Status.State != nil && *sg.Raw().Status.State == StateInCreation {
			return fmt.Errorf("cannot update security group while it is in 'InCreation' state. Please wait until the security group is fully created")
		}

		if name != "" {
			sg.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			sg.RetaggedAs(tags...)
		}

		updated, err := client.FromNetwork().SecurityGroups().Update(ctx, sg)
		if err != nil {
			return fmt.Errorf("updating security group: %w", apiErrFromV2(err))
		}
		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			updateRegion := ""
			if raw.Metadata.LocationResponse != nil {
				updateRegion = string(raw.Metadata.LocationResponse.Value)
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			nameVal := ""
			if raw.Metadata.Name != nil {
				nameVal = *raw.Metadata.Name
			}
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			row := []string{nameVal, id, updateRegion, status}
			PrintOutput(raw, headers, [][]string{row})
		} else {
			fmt.Println(msgUpdatedAsync("Security group", sgID))
		}
		return nil
	},
}

var securitygroupDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [securitygroup-id]",
	Short: "Delete a security group",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcID := args[0]
		sgID := args[1]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("security group", sgID)
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
			_, err = client.FromNetwork().SecurityGroups().Get(ctx, aruba.SecurityGroupRef(projectID, vpcID, sgID))
			if err != nil {
				return fmt.Errorf("dry-run: security group not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("security group", sgID))
			return nil
		}

		err = client.FromNetwork().SecurityGroups().Delete(ctx, aruba.SecurityGroupRef(projectID, vpcID, sgID))
		if err != nil {
			return fmt.Errorf("deleting security group: %w", apiErrFromV2(err))
		}
		headers := []TableColumn{
			{Header: "ID", Width: 26},
			{Header: "STATUS", Width: 15},
		}
		status := "deleted"
		result := struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}{sgID, status}
		PrintOutput(result, headers, [][]string{{sgID, status}})
		return nil
	},
}
