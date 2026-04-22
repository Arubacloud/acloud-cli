package cmd

import (
	"fmt"

	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

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
		req := types.SecurityGroupRequest{
			Metadata: types.ResourceMetadataRequest{
				Name: name,
				Tags: tags,
			},
		}
		resp, err := client.FromNetwork().SecurityGroups().Create(ctx, projectID, vpcID, req, nil)
		if err != nil {
			return fmt.Errorf("creating security group: %w", err)
		}
		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}
		if resp != nil && resp.Data != nil && resp.Data.Metadata.ID != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			region := ""
			if resp.Data.Metadata.LocationResponse != nil {
				region = resp.Data.Metadata.LocationResponse.Value
			}
			row := []string{
				name,
				*resp.Data.Metadata.ID,
				region,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					} else {
						return ""
					}
				}(),
			}
			PrintOutput(resp.Data, headers, [][]string{row})
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
		resp, err := client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
		if err != nil {
			return fmt.Errorf("getting security group: %w", err)
		}
		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}
		if resp != nil && resp.Data != nil {
			sg := resp.Data
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
				if sg.Metadata.LocationResponse != nil {
					fmt.Printf("Region:          %s\n", sg.Metadata.LocationResponse.Value)
				}
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
		resp, err := client.FromNetwork().SecurityGroups().List(ctx, projectID, vpcID, listParams(cmd))
		if err != nil {
			return fmt.Errorf("listing security groups: %w", err)
		}
		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}
		if resp != nil && resp.Data != nil && len(resp.Data.Values) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			var rows [][]string
			for _, sg := range resp.Data.Values {
				name := ""
				if sg.Metadata.Name != nil {
					name = *sg.Metadata.Name
				}
				id := ""
				if sg.Metadata.ID != nil {
					id = *sg.Metadata.ID
				}
				region := ""
				if sg.Metadata.LocationResponse != nil {
					region = sg.Metadata.LocationResponse.Value
				}
				status := ""
				if sg.Status.State != nil {
					status = *sg.Status.State
				}
				rows = append(rows, []string{name, id, region, status})
			}
			PrintOutput(resp.Data, headers, rows)
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
		// Fetch current security group details
		getResp, err := client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			return fmt.Errorf("fetching current security group: %w", err)
		}
		current := getResp.Data
		// Block update if security group is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == StateInCreation {
			return fmt.Errorf("cannot update security group while it is in 'InCreation' state. Please wait until the security group is fully created")
		}
		// Get region value
		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			return fmt.Errorf("unable to determine region value for security group")
		}
		// Build update request by merging user input with all current valid fields
		req := types.SecurityGroupRequest{
			Metadata: types.ResourceMetadataRequest{
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
		}
		resp, err := client.FromNetwork().SecurityGroups().Update(ctx, projectID, vpcID, sgID, req, nil)
		if err != nil {
			return fmt.Errorf("updating security group: %w", err)
		}
		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
		}
		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			updateRegion := ""
			if resp.Data.Metadata.LocationResponse != nil {
				updateRegion = resp.Data.Metadata.LocationResponse.Value
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
				updateRegion,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintOutput(resp.Data, headers, [][]string{row})
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

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
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
			_, err = client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
			if err != nil {
				return fmt.Errorf("dry-run: security group not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("security group", sgID))
			return nil
		}

		resp, err := client.FromNetwork().SecurityGroups().Delete(ctx, projectID, vpcID, sgID, nil)
		if err != nil {
			return fmt.Errorf("deleting security group: %w", err)
		}
		if resp != nil && resp.IsError() {
			return apiErrFromResp(resp.StatusCode, resp.Error)
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
