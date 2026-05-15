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

// vpcRef builds the combined project+VPC Ref used by Get/Delete/Update; also
// used as the parent Ref for VPC-scoped resources (subnet, securitygroup, vpcpeering).
func vpcRef(projectID, vpcID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID + "/providers/Aruba.Network/vpcs/" + vpcID)
}

func vpcListPayload(l *aruba.List[*aruba.VPC]) any {
	if r, ok := l.Raw().(*types.Response[types.VPCList]); ok && r != nil {
		return r.Data
	}
	return nil
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
	list, err := client.FromNetwork().VPCs().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, vpc := range list.Items() {
			id := vpc.VPCID()
			name := vpc.Name()
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
	RunE: func(cmd *cobra.Command, args []string) error {
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
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			NotDefault().
			WithPreset(false)
		if len(tags) > 0 {
			vpc.ReplaceTags(tags...)
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromNetwork().VPCs().Create(ctx, vpc)
		if err != nil {
			return fmt.Errorf("creating VPC: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil {
			fmt.Printf("\n%s\n", msgCreated("VPC", name))
			if r.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *r.Metadata.ID)
			}
			if r.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *r.Metadata.Name)
			}
			fmt.Printf("Default: %t\n", r.Properties.Default)
			if len(r.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", r.Metadata.Tags)
			}
		} else {
			fmt.Println(msgCreatedAsync("VPC", name))
		}
		return nil
	},
}

var vpcGetCmd = &cobra.Command{
	Use:   "get <vpc-id>",
	Short: "Get VPC details",
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
		got, err := client.FromNetwork().VPCs().Get(ctx, vpcRef(projectID, vpcID))
		if err != nil {
			return fmt.Errorf("getting VPC details: %w", apiErrFromV2(err))
		}

		vpc := got.Raw()
		if vpc != nil {
			fmt.Println("\nVPC Details:")
			fmt.Println("============")

			if vpc.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *vpc.Metadata.ID)
			}
			if vpc.Metadata.URI != nil {
				fmt.Printf("URI:             %s\n", *vpc.Metadata.URI)
			}
			if vpc.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *vpc.Metadata.Name)
			}
			if vpc.Metadata.LocationResponse != nil && vpc.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", vpc.Metadata.LocationResponse.Value)
			}
			fmt.Printf("Default:         %t\n", vpc.Properties.Default)
			fmt.Printf("Linked Resources: %d\n", len(vpc.Properties.LinkedResources))

			if vpc.Metadata.CreationDate != nil {
				fmt.Printf("Creation Date:   %s\n", vpc.Metadata.CreationDate.Format(DateLayout))
			}
			if vpc.Metadata.CreatedBy != nil {
				fmt.Printf("Created By:      %s\n", *vpc.Metadata.CreatedBy)
			}

			if len(vpc.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", vpc.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			if vpc.Status.State != nil {
				fmt.Printf("Status:          %s\n", *vpc.Status.State)
			}
		} else {
			fmt.Println("VPC not found or no data returned.")
		}
		return nil
	},
}

var vpcUpdateCmd = &cobra.Command{
	Use:   "update <vpc-id>",
	Short: "Update a VPC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
		cur, err := client.FromNetwork().VPCs().Get(ctx, vpcRef(projectID, vpcID))
		if err != nil {
			return fmt.Errorf("getting VPC details: %w", apiErrFromV2(err))
		}
		// SDK VPC.Get does not backfill projectID from the Ref (unlike other resources).
		cur.IntoProject(projectRef(projectID))

		r := cur.Raw()
		if r == nil {
			return fmt.Errorf("VPC not found")
		}

		if r.Status.State != nil && *r.Status.State == StateInCreation {
			return fmt.Errorf("cannot update VPC while it is in 'InCreation' state. Please wait until the VPC is fully created")
		}

		if cur.Region() == "" {
			return fmt.Errorf("unable to determine region value for VPC")
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}

		updated, err := client.FromNetwork().VPCs().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating VPC: %w", apiErrFromV2(err))
		}

		ur := updated.Raw()
		if ur != nil {
			fmt.Printf("\n%s\n", msgUpdated("VPC", vpcID))
			if ur.Metadata.ID != nil {
				fmt.Printf("ID:      %s\n", *ur.Metadata.ID)
			}
			if ur.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *ur.Metadata.Name)
			}
			if len(ur.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", ur.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("VPC", vpcID))
		}
		return nil
	},
}

var vpcDeleteCmd = &cobra.Command{
	Use:   "delete <vpc-id>",
	Short: "Delete a VPC",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
			_, err = client.FromNetwork().VPCs().Get(ctx, vpcRef(projectID, vpcID))
			if err != nil {
				return fmt.Errorf("dry-run: VPC not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("VPC", vpcID))
			return nil
		}

		if err := client.FromNetwork().VPCs().Delete(ctx, vpcRef(projectID, vpcID)); err != nil {
			return fmt.Errorf("deleting VPC: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("VPC", vpcID))
		return nil
	},
}

var vpcListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPCs",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		list, err := client.FromNetwork().VPCs().List(ctx, projectRef(projectID), listOpts(cmd)...)
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
				r := vpc.Raw()
				name := vpc.Name()
				id := vpc.VPCID()

				region := ""
				subnets := "0"
				status := ""
				if r != nil {
					if r.Metadata.LocationResponse != nil {
						region = string(r.Metadata.LocationResponse.Value)
					}
					subnets = fmt.Sprintf("%d", len(r.Properties.LinkedResources))
					if r.Status.State != nil {
						status = *r.Status.State
					}
				}

				rows = append(rows, []string{name, id, region, subnets, status})
			}

			PrintOutput(vpcListPayload(list), headers, rows)
		} else {
			fmt.Println("No VPCs found")
		}
		return nil
	},
}
