package cmd

import (
	"context"
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
	securitygroupCreateCmd.Flags().String("name", "", "Security group name (required)")
	securitygroupCreateCmd.Flags().String("region", "", "Region code (required)")
	securitygroupCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	securitygroupDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if name == "" || region == "" {
			fmt.Println("Error: --name and --region are required")
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
		req := types.SecurityGroupRequest{
			Metadata: types.ResourceMetadataRequest{
				Name: name,
				Tags: tags,
			},
		}
		resp, err := client.FromNetwork().SecurityGroups().Create(ctx, projectID, vpcID, req, nil)
		if err != nil {
			fmt.Printf("Error creating security group: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to create security group - Status: %d\n", resp.StatusCode)
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
				{Header: "REGION", Width: 18},
				{Header: "STATUS", Width: 15},
			}
			row := []string{
				name,
				*resp.Data.Metadata.ID,
				resp.Data.Metadata.LocationResponse.Value,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					} else {
						return ""
					}
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Println("Security group created, but no ID returned.")
		}
	},
}

var securitygroupGetCmd = &cobra.Command{
	Use:   "get [vpc-id] [securitygroup-id]",
	Short: "Get security group details",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		sgID := args[1]
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
		resp, err := client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
		if err != nil {
			fmt.Printf("Error getting security group: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to get security group - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
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
				fmt.Printf("Region:          %s\n", sg.Metadata.LocationResponse.Value)
			}
			if sg.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", sg.Metadata.CreationDate.Format("02-01-2006 15:04:05"))
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
	},
}

var securitygroupListCmd = &cobra.Command{
	Use:   "list [vpc-id]",
	Short: "List security groups for a VPC",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
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
		resp, err := client.FromNetwork().SecurityGroups().List(ctx, projectID, vpcID, nil)
		if err != nil {
			fmt.Printf("Error listing security groups: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to list security groups - Status: %d\n", resp.StatusCode)
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
			PrintTable(headers, rows)
		} else {
			fmt.Println("No security groups found.")
		}
	},
}

var securitygroupUpdateCmd = &cobra.Command{
	Use:   "update [vpc-id] [securitygroup-id]",
	Short: "Update a security group",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		sgID := args[1]
		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		if name == "" && !cmd.Flags().Changed("tags") {
			fmt.Println("Error: at least one of --name or --tags must be provided")
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
		// Fetch current security group details
		getResp, err := client.FromNetwork().SecurityGroups().Get(ctx, projectID, vpcID, sgID, nil)
		if err != nil || getResp == nil || getResp.Data == nil {
			fmt.Printf("Error fetching current security group: %v\n", err)
			return
		}
		current := getResp.Data
		// Block update if security group is in 'InCreation' state
		if current.Status.State != nil && *current.Status.State == "InCreation" {
			fmt.Println("Error: Cannot update security group while it is in 'InCreation' state. Please wait until the security group is fully created.")
			return
		}
		// Get region value
		regionValue := ""
		if current.Metadata.LocationResponse != nil {
			regionValue = current.Metadata.LocationResponse.Value
		}
		if regionValue == "" {
			fmt.Println("Error: Unable to determine region value for security group")
			return
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
			fmt.Printf("Error updating security group: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to update security group - Status: %d\n", resp.StatusCode)
			if resp.Error.Title != nil {
				fmt.Printf("Error: %s\n", *resp.Error.Title)
			}
			if resp.Error.Detail != nil {
				fmt.Printf("Detail: %s\n", *resp.Error.Detail)
			}
			return
		}
		if resp != nil && resp.Data != nil {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 26},
				{Header: "REGION", Width: 18},
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
				resp.Data.Metadata.LocationResponse.Value,
				func() string {
					if resp.Data.Status.State != nil {
						return *resp.Data.Status.State
					}
					return ""
				}(),
			}
			PrintTable(headers, [][]string{row})
		} else {
			fmt.Printf("Security group '%s' updated.\n", sgID)
		}
	},
}

var securitygroupDeleteCmd = &cobra.Command{
	Use:   "delete [vpc-id] [securitygroup-id]",
	Short: "Delete a security group",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		vpcID := args[0]
		sgID := args[1]

		// Get skip confirmation flag
		skipConfirm, _ := cmd.Flags().GetBool("yes")

		// Prompt for confirmation unless --yes flag is used
		if !skipConfirm {
			fmt.Printf("Are you sure you want to delete security group %s? This action cannot be undone.\n", sgID)
			fmt.Print("Type 'yes' to confirm: ")
			var response string
			fmt.Scanln(&response)
			if response != "yes" && response != "y" {
				fmt.Println("Delete cancelled")
				return
			}
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
		resp, err := client.FromNetwork().SecurityGroups().Delete(ctx, projectID, vpcID, sgID, nil)
		if err != nil {
			fmt.Printf("Error deleting security group: %v\n", err)
			return
		}
		if resp != nil && resp.IsError() && resp.Error != nil {
			fmt.Printf("Failed to delete security group - Status: %d\n", resp.StatusCode)
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
		PrintTable(headers, [][]string{{sgID, status}})
	},
}
