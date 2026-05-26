package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/spf13/cobra"
)

func init() {
	databaseCmd.AddCommand(dbaasCmd)
	dbaasCmd.AddCommand(dbaasCreateCmd)
	dbaasCmd.AddCommand(dbaasGetCmd)
	dbaasCmd.AddCommand(dbaasUpdateCmd)
	dbaasCmd.AddCommand(dbaasDeleteCmd)
	dbaasCmd.AddCommand(dbaasListCmd)

	dbaasCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasCreateCmd.Flags().String("name", "", "Name for the DBaaS instance (required)")
	dbaasCreateCmd.Flags().String("region", "", "Region code (required)")
	dbaasCreateCmd.Flags().String("zone", "", "Availability zone (required, e.g. ITBG-1)")
	dbaasCreateCmd.Flags().String("engine-id", "", "Database engine ID (required, e.g. mysql-8.0)")
	dbaasCreateCmd.Flags().String("flavor", "", "DBaaS flavor name (required, e.g. DBO4A8)")
	dbaasCreateCmd.Flags().Int("storage-size", 0, "Storage size in GB (required)")
	dbaasCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	dbaasCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required when project has a VPC)")
	dbaasCreateCmd.Flags().String("subnet-uri", "", "Subnet URI (required when project has a VPC)")
	dbaasCreateCmd.Flags().String("security-group-uri", "", "Security group URI (required when project has a VPC)")
	dbaasCreateCmd.Flags().String("elastic-ip-uri", "", "Elastic IP URI (optional)")
	dbaasCreateCmd.MarkFlagRequired("name")
	dbaasCreateCmd.MarkFlagRequired("region")
	dbaasCreateCmd.MarkFlagRequired("zone")
	dbaasCreateCmd.MarkFlagRequired("engine-id")
	dbaasCreateCmd.MarkFlagRequired("flavor")
	dbaasCreateCmd.MarkFlagRequired("storage-size")

	dbaasGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	dbaasUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasUpdateCmd.Flags().String("name", "", "New name for the DBaaS instance")
	dbaasUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	dbaasDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	dbaasDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	dbaasListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	dbaasListCmd.Flags().Int("limit", 0, "Maximum number of results to return (0 = no limit)")
	dbaasListCmd.Flags().Int("offset", 0, "Number of results to skip")

	dbaasGetCmd.ValidArgsFunction = completeDBaaSID
	dbaasUpdateCmd.ValidArgsFunction = completeDBaaSID
	dbaasDeleteCmd.ValidArgsFunction = completeDBaaSID
}

func completeDBaaSID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromDatabase().DBaaS().List(ctx, aruba.URI("/projects/"+projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, d := range list.Items() {
			raw := d.Raw()
			if raw != nil && raw.Metadata.ID != nil && raw.Metadata.Name != nil {
				id := *raw.Metadata.ID
				if toComplete == "" || strings.HasPrefix(id, toComplete) {
					completions = append(completions, fmt.Sprintf("%s\t%s", id, *raw.Metadata.Name))
				}
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

var dbaasCmd = &cobra.Command{
	Use:   "dbaas",
	Short: "Manage DBaaS resources",
	Long:  `Perform CRUD operations on DBaaS resources in Aruba Cloud.`,
}

var dbaasCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new DBaaS instance",
	Long: `Create a new managed database instance in the specified region.

Use --engine-id to select the database engine (e.g., MySQL, PostgreSQL).
Use --flavor to select the compute profile (CPU/RAM).

After creation, add databases with 'acloud database dbaas database create'
and users with 'acloud database dbaas user create'.`,
	Example: `  acloud database dbaas create \
    --name my-db --region IT-BG \
    --engine-id <engine-id> \
    --flavor <flavor-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		zone, _ := cmd.Flags().GetString("zone")
		engineID, _ := cmd.Flags().GetString("engine-id")
		flavor, _ := cmd.Flags().GetString("flavor")
		storageSize, _ := cmd.Flags().GetInt("storage-size")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURI, _ := cmd.Flags().GetString("subnet-uri")
		sgURI, _ := cmd.Flags().GetString("security-group-uri")
		elasticIPURI, _ := cmd.Flags().GetString("elastic-ip-uri")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		dbaas := aruba.NewDBaaS().
			IntoProject(aruba.URI("/projects/" + projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			OfEngine(aruba.DatabaseEngine(engineID)).
			OfFlavor(aruba.DBaaSFlavor(flavor)).
			ReplaceTags(tags...)

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromDatabase().DBaaS().Create(ctx, dbaas)
		if err != nil {
			return fmt.Errorf("creating DBaaS instance: %w", apiErrFromV2(err))
		}

		if created != nil && created.Raw() != nil {
			raw := created.Raw()
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "ENGINE", Width: 20},
				{Header: "VERSION", Width: 15},
				{Header: "FLAVOR", Width: 20},
				{Header: "REGION", Width: 20},
			}
			id := ""
			if raw.Metadata.ID != nil {
				id = *raw.Metadata.ID
			}
			nameVal := ""
			if raw.Metadata.Name != nil {
				nameVal = *raw.Metadata.Name
			}
			engine := ""
			if raw.Properties.Engine != nil && raw.Properties.Engine.Type != nil {
				engine = *raw.Properties.Engine.Type
			}
			version := ""
			if raw.Properties.Engine != nil && raw.Properties.Engine.Version != nil {
				version = *raw.Properties.Engine.Version
			}
			flavorVal := ""
			if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
				flavorVal = *raw.Properties.Flavor.Name
			}
			regionVal := ""
			if raw.Metadata.LocationResponse != nil {
				regionVal = string(raw.Metadata.LocationResponse.Value)
			}
			PrintOutput(raw, headers, [][]string{{id, nameVal, engine, version, flavorVal, regionVal}})
		} else {
			fmt.Println(msgCreatedAsync("DBaaS instance", name))
		}
		return nil
	},
}

var dbaasGetCmd = &cobra.Command{
	Use:   "get [dbaas-id]",
	Short: "Get DBaaS instance details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

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
		dbaas, err := client.FromDatabase().DBaaS().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/dbaas/"+dbaasID))
		if err != nil {
			return fmt.Errorf("getting DBaaS instance: %w", apiErrFromV2(err))
		}

		if dbaas != nil && dbaas.Raw() != nil {
			raw := dbaas.Raw()

			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(raw, nil, nil)
				return nil
			}

			fmt.Println("\nDBaaS Instance Details:")
			fmt.Println("=======================")
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
			if raw.Properties.Engine != nil {
				if raw.Properties.Engine.Type != nil {
					fmt.Printf("Engine Type:     %s\n", *raw.Properties.Engine.Type)
				}
				if raw.Properties.Engine.Version != nil {
					fmt.Printf("Engine Version:  %s\n", *raw.Properties.Engine.Version)
				}
				if raw.Properties.Engine.Name != nil {
					fmt.Printf("Engine Name:     %s\n", *raw.Properties.Engine.Name)
				}
			}
			if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
				fmt.Printf("Flavor:          %s\n", *raw.Properties.Flavor.Name)
			}
			if raw.Status.State != nil {
				fmt.Printf("Status:          %s\n", string(*raw.Status.State))
			}
			if raw.Metadata.CreationDate != nil && !raw.Metadata.CreationDate.IsZero() {
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
			fmt.Println()
		} else {
			fmt.Println("DBaaS instance not found")
		}
		return nil
	},
}

var dbaasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all DBaaS instances",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		list, err := client.FromDatabase().DBaaS().List(ctx, aruba.URI("/projects/"+projectID))
		if err != nil {
			return fmt.Errorf("listing DBaaS instances: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 30},
				{Header: "ID", Width: 30},
				{Header: "ENGINE", Width: 20},
				{Header: "VERSION", Width: 15},
				{Header: "FLAVOR", Width: 20},
				{Header: "REGION", Width: 20},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, d := range list.Items() {
				raw := d.Raw()
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
				engine := ""
				if raw.Properties.Engine != nil && raw.Properties.Engine.Type != nil {
					engine = *raw.Properties.Engine.Type
				}
				version := ""
				if raw.Properties.Engine != nil && raw.Properties.Engine.Version != nil {
					version = *raw.Properties.Engine.Version
				}
				flavor := ""
				if raw.Properties.Flavor != nil && raw.Properties.Flavor.Name != nil {
					flavor = *raw.Properties.Flavor.Name
				}
				region := ""
				if raw.Metadata.LocationResponse != nil {
					region = string(raw.Metadata.LocationResponse.Value)
				}
				status := ""
				if raw.Status.State != nil {
					status = string(*raw.Status.State)
				}
				rows = append(rows, []string{name, id, engine, version, flavor, region, status})
			}
			PrintOutput(list.Raw(), headers, rows)
		} else {
			fmt.Println("No DBaaS instances found")
		}
		return nil
	},
}

var dbaasUpdateCmd = &cobra.Command{
	Use:   "update [dbaas-id]",
	Short: "Update a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		tags, _ := cmd.Flags().GetStringSlice("tags")

		if name == "" && !cmd.Flags().Changed("tags") {
			return fmt.Errorf("at least one of --name or --tags must be provided")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		dbaas, err := client.FromDatabase().DBaaS().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/dbaas/"+dbaasID))
		if err != nil {
			return fmt.Errorf("getting DBaaS instance: %w", err)
		}
		if dbaas == nil || dbaas.Raw() == nil {
			return fmt.Errorf("DBaaS instance not found")
		}

		if name != "" {
			dbaas.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			dbaas.ReplaceTags(tags...)
		}

		updated, err := client.FromDatabase().DBaaS().Update(ctx, dbaas)
		if err != nil {
			return fmt.Errorf("updating DBaaS instance: %w", apiErrFromV2(err))
		}

		if updated != nil && updated.Raw() != nil {
			raw := updated.Raw()
			fmt.Printf("\n%s\n", msgUpdated("DBaaS instance", dbaasID))
			if raw.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *raw.Metadata.ID)
			}
			if raw.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *raw.Metadata.Name)
			}
			if len(raw.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", raw.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("DBaaS instance", dbaasID))
		}
		return nil
	},
}

var dbaasDeleteCmd = &cobra.Command{
	Use:   "delete [dbaas-id]",
	Short: "Delete a DBaaS instance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbaasID := args[0]

		confirm, _ := cmd.Flags().GetBool("yes")
		if !confirm {
			ok, err := confirmDelete("DBaaS instance", dbaasID)
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
			_, err = client.FromDatabase().DBaaS().Get(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/dbaas/"+dbaasID))
			if err != nil {
				return fmt.Errorf("dry-run: DBaaS instance not found or inaccessible: %w", err)
			}
			fmt.Println(msgDryRun("DBaaS instance", dbaasID))
			return nil
		}

		err = client.FromDatabase().DBaaS().Delete(ctx, aruba.URI("/projects/"+projectID+"/providers/Aruba.Database/dbaas/"+dbaasID))
		if err != nil {
			return fmt.Errorf("deleting DBaaS instance: %w", err)
		}

		fmt.Println(msgDeleted("DBaaS instance", dbaasID))
		return nil
	},
}
