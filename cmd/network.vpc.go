package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {

	// VPC
	networkCmd.AddCommand(vpcCmd)
	vpcCmd.AddCommand(vpcCreateCmd)
	vpcCmd.AddCommand(vpcGetCmd)
	vpcCmd.AddCommand(vpcUpdateCmd)
	vpcCmd.AddCommand(vpcDeleteCmd)
	vpcCmd.AddCommand(vpcListCmd)

	// Add flags for VPC commands
	vpcCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcCreateCmd.Flags().String("name", "", "Name for the VPC")
	vpcCreateCmd.Flags().String("region", "", "Region code (e.g., IT-BG)")
	vpcCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	vpcCreateCmd.MarkFlagRequired("name")
	vpcCreateCmd.MarkFlagRequired("region")
	vpcGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcUpdateCmd.Flags().String("name", "", "New name for the VPC")
	vpcUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")
	vpcDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	vpcDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")
	vpcListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	vpcListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	vpcListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	// Set up auto-completion for resource IDs
	vpcGetCmd.ValidArgsFunction = completeVPCID
	vpcUpdateCmd.ValidArgsFunction = completeVPCID
	vpcDeleteCmd.ValidArgsFunction = completeVPCID
}

// Completion functions for network resources

func completeVPCID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromNetwork().VPCs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vpc := range list.Items() {
			id := vpc.ID()
			if id != "" && (toComplete == "" || strings.HasPrefix(id, toComplete)) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, vpc.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// VPC subcommands
var vpcCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Manage VPCs",
	Long:  `Perform CRUD operations on VPCs in Aruba Cloud.`,
}

var vpcCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new VPC",
	Long: `Create a new Virtual Private Cloud (VPC) in the specified region.

A VPC is the top-level network boundary. Subnets, security groups, and other
network resources are created within a VPC.`,
	Example: `  acloud network vpc create --name my-vpc --region IT-BG
  acloud network vpc create --name prod-vpc --region IT-BG --tags env=prod,team=infra`,
	Args: cobra.NoArgs,
	RunE: runVPCCreate,
}

var vpcGetCmd = &cobra.Command{
	Use:   "get <vpc-id>",
	Short: "Get VPC details",
	Args:  cobra.ExactArgs(1),
	RunE:  runVPCGet,
}

var vpcUpdateCmd = &cobra.Command{
	Use:   "update <vpc-id>",
	Short: "Update a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  runVPCUpdate,
}

var vpcDeleteCmd = &cobra.Command{
	Use:   "delete <vpc-id>",
	Short: "Delete a VPC",
	Args:  cobra.ExactArgs(1),
	RunE:  runVPCDelete,
}

var vpcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPCs",
	Args:  cobra.NoArgs,
	RunE:  runVPCList,
}

func runVPCCreate(cmd *cobra.Command, args []string) error {
	name, _ := cmd.Flags().GetString("name")
	region, _ := cmd.Flags().GetString("region")
	tags, _ := cmd.Flags().GetStringSlice("tags")

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return err
	}

	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	vpc := aruba.NewVPC().
		InProject(aruba.URI("/projects/" + projectID)).
		Named(name).
		InRegion(aruba.Region(region)).
		RetaggedAs(tags...)

	ctx, cancel := newCtx()
	defer cancel()
	created, err := client.FromNetwork().VPCs().Create(ctx, vpc)
	if err != nil {
		return fmt.Errorf("creating VPC: %w", apiErrFromV2(err))
	}

	if created != nil && created.Raw() != nil {
		raw := created.Raw()
		fmt.Printf("\n%s\n", msgCreated("VPC", name))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		fmt.Printf("Default: %t\n", raw.Properties.Default)
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgCreatedAsync("VPC", name))
	}
	return nil
}

func runVPCGet(cmd *cobra.Command, args []string) error {
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
	vpc, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(projectID, vpcID))
	if err != nil {
		return fmt.Errorf("getting VPC details: %w", apiErrFromV2(err))
	}

	if vpc != nil && vpc.Raw() != nil {
		raw := vpc.Raw()

		fmt.Println("\nVPC Details:")
		fmt.Println("============")

		if raw.Metadata.ID != nil {
			fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.URI != nil {
			fmt.Printf("URI:             %s\n", *raw.Metadata.URI)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
		}
		if raw.Metadata.LocationResponse != nil && raw.Metadata.LocationResponse.Value != "" {
			fmt.Printf("Region:          %s\n", raw.Metadata.LocationResponse.Value)
		}
		fmt.Printf("Default:         %t\n", raw.Properties.Default)
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
	} else {
		fmt.Println("VPC not found or no data returned.")
	}
	return nil
}

func runVPCUpdate(cmd *cobra.Command, args []string) error {
	vpcID := args[0]

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
	vpc, err := client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(projectID, vpcID))
	if err != nil {
		return fmt.Errorf("getting VPC details: %w", apiErrFromV2(err))
	}

	if vpc == nil || vpc.Raw() == nil {
		return fmt.Errorf("VPC not found")
	}

	if vpc.Raw().Status.State != nil && *vpc.Raw().Status.State == StateInCreation {
		return fmt.Errorf("cannot update VPC while it is in 'InCreation' state. Please wait until the VPC is fully created")
	}

	if name != "" {
		vpc.Named(name)
	}
	if len(tags) > 0 {
		vpc.RetaggedAs(tags...)
	}

	updated, err := client.FromNetwork().VPCs().Update(ctx, vpc)
	if err != nil {
		return fmt.Errorf("updating VPC: %w", apiErrFromV2(err))
	}

	if updated != nil && updated.Raw() != nil {
		raw := updated.Raw()
		fmt.Printf("\n%s\n", msgUpdated("VPC", vpcID))
		if raw.Metadata.ID != nil {
			fmt.Printf("ID:      %s\n", *raw.Metadata.ID)
		}
		if raw.Metadata.Name != nil {
			fmt.Printf("Name:    %s\n", *raw.Metadata.Name)
		}
		if len(raw.Metadata.Tags) > 0 {
			fmt.Printf("Tags:    %v\n", raw.Metadata.Tags)
		}
	} else {
		fmt.Println(msgUpdatedAsync("VPC", vpcID))
	}
	return nil
}

func runVPCDelete(cmd *cobra.Command, args []string) error {
	vpcID := args[0]

	skipConfirm, _ := cmd.Flags().GetBool("yes")

	if !skipConfirm {
		ok, err := confirmDelete("VPC", vpcID)
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
		_, err = client.FromNetwork().VPCs().Get(ctx, aruba.VPCRef(projectID, vpcID))
		if err != nil {
			return fmt.Errorf("dry-run: VPC not found or inaccessible: %w", apiErrFromV2(err))
		}
		fmt.Println(msgDryRun("VPC", vpcID))
		return nil
	}

	err = client.FromNetwork().VPCs().Delete(ctx, aruba.VPCRef(projectID, vpcID))
	if err != nil {
		return fmt.Errorf("deleting VPC: %w", apiErrFromV2(err))
	}

	fmt.Println(msgDeleted("VPC", vpcID))
	return nil
}

func runVPCList(cmd *cobra.Command, args []string) error {
	client, err := GetArubaClient()
	if err != nil {
		return fmt.Errorf("initializing client: %w", err)
	}

	projectID, err := GetProjectID(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := newCtx()
	defer cancel()
	list, err := client.FromNetwork().VPCs().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return fmt.Errorf("listing VPCs: %w", apiErrFromV2(err))
	}

	if list != nil && len(list.Items()) > 0 {
		headers := []TableColumn{
			{Header: "NAME", Width: 40},
			{Header: "ID", Width: 25},
			{Header: "REGION", Width: 18},
			{Header: "SUBNETS", Width: 10},
			{Header: "STATUS", Width: 15},
		}

		var rows [][]string
		for _, vpc := range list.Items() {
			raw := vpc.Raw()
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
			subnets := fmt.Sprintf("%d", len(raw.Properties.LinkedResources))
			status := ""
			if raw.Status.State != nil {
				status = string(*raw.Status.State)
			}
			rows = append(rows, []string{name, id, region, subnets, status})
		}

		PrintOutput(list, headers, rows)
	} else {
		fmt.Println("No VPCs found")
	}
	return nil
}
