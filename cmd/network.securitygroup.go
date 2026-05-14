package cmd

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

// securityGroupRef is shared with securityrule.go and subnet.go (via vpcRef).
func securityGroupRef(projectID, vpcID, sgID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/vpcs/" + vpcID + "/securitygroups/" + sgID)
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
		_, _ = cmd.Flags().GetString("region") // required by Cobra, not used in SDK request
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
			IntoVPC(vpcRef(projectID, vpcID)).
			Named(name)
		if len(tags) > 0 {
			sg.ReplaceTags(tags...)
		}

		created, err := client.FromNetwork().SecurityGroups().Create(ctx, sg)
		if err != nil {
			return fmt.Errorf("creating security group: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil && r.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			region := ""
			if r.Metadata.LocationResponse != nil {
				region = string(r.Metadata.LocationResponse.Value)
			}
			status := ""
			if r.Status.State != nil {
				status = *r.Status.State
			}
			PrintOutput(r, headers, [][]string{{name, *r.Metadata.ID, region, status}})
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
		got, err := client.FromNetwork().SecurityGroups().Get(ctx, securityGroupRef(projectID, vpcID, sgID))
		if err != nil {
			return fmt.Errorf("getting security group: %w", apiErrFromV2(err))
		}

		sg := got.Raw()
		if sg != nil {
			fmt.Println("\nSecurity Group Details:")
			fmt.Println("======================")
			if sg.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *sg.Metadata.ID)
			}
			if sg.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *sg.Metadata.URI)
			}
			if sg.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *sg.Metadata.Name)
			}
			if sg.Metadata.LocationResponse != nil {
				fmt.Printf("Region:          %s\n", sg.Metadata.LocationResponse.Value)
			}
			if sg.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", sg.Metadata.CreationDate.Format(DateLayout))
			}
			if sg.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *sg.Metadata.CreatedBy)
			}
			if len(sg.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", sg.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}
			if sg.Status.State != nil {
				fmt.Printf("Status:          %s\n", *sg.Status.State)
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
		list, err := client.FromNetwork().SecurityGroups().List(ctx, vpcRef(projectID, vpcID), listOpts(cmd)...)
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
				r := sg.Raw()
				name := sg.Name()
				id := sg.SecurityGroupID()
				region := ""
				status := ""
				if r != nil {
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}
				rows = append(rows, []string{name, id, region, status})
			}
			PrintOutput(securityGroupListPayload(list), headers, rows)
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

		cur, err := client.FromNetwork().SecurityGroups().Get(ctx, securityGroupRef(projectID, vpcID, sgID))
		if err != nil {
			return fmt.Errorf("fetching current security group: %w", apiErrFromV2(err))
		}
		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("security group not found")
		}
		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update security group while it is in 'InCreation' state. Please wait until the security group is fully created")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().SecurityGroups().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating security group: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
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
			updRegion := ""
			if ur.Metadata.LocationResponse != nil {
				updRegion = string(ur.Metadata.LocationResponse.Value)
			}
			updStatus := ""
			if ur.Status.State != nil {
				updStatus = *ur.Status.State
			}
			PrintOutput(ur, headers, [][]string{{updName, updID, updRegion, updStatus}})
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
			_, err = client.FromNetwork().SecurityGroups().Get(ctx, securityGroupRef(projectID, vpcID, sgID))
			if err != nil {
				return fmt.Errorf("dry-run: security group not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("security group", sgID))
			return nil
		}

		if err := client.FromNetwork().SecurityGroups().Delete(ctx, securityGroupRef(projectID, vpcID, sgID)); err != nil {
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
