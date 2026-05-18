package cmd

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Arubacloud/sdk-go/pkg/aruba"
	"github.com/Arubacloud/sdk-go/pkg/types"
	"github.com/spf13/cobra"
)

func init() {
	cloudserverCreateCmd.Flags().String("boot-disk-uri", "", "Bootable block storage URI (required, e.g., /projects/{project-id}/providers/Aruba.Storage/blockStorages/{volume-id})")
	cloudserverCreateCmd.MarkFlagRequired("boot-disk-uri")
	cloudserverCreateCmd.MarkFlagRequired("vpc-uri")
	cloudserverCreateCmd.MarkFlagRequired("subnet-uri")
	cloudserverCreateCmd.MarkFlagRequired("security-group-uri")
	// CloudServer commands
	computeCmd.AddCommand(cloudserverCmd)
	cloudserverCmd.AddCommand(cloudserverCreateCmd)
	cloudserverCmd.AddCommand(cloudserverGetCmd)
	cloudserverCmd.AddCommand(cloudserverUpdateCmd)
	cloudserverCmd.AddCommand(cloudserverDeleteCmd)
	cloudserverCmd.AddCommand(cloudserverListCmd)
	cloudserverCmd.AddCommand(cloudserverPowerOnCmd)
	cloudserverCmd.AddCommand(cloudserverPowerOffCmd)
	cloudserverCmd.AddCommand(cloudserverSetPasswordCmd)
	cloudserverCmd.AddCommand(cloudserverConnectCmd)

	// Add flags for cloudserver commands
	cloudserverCreateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverCreateCmd.Flags().String("name", "", "Name for the cloud server (required)")
	cloudserverCreateCmd.Flags().String("region", "", "Region code (required)")
	cloudserverCreateCmd.Flags().String("zone", "", "Zone code (required, e.g., itbg1-a)")
	cloudserverCreateCmd.Flags().String("flavor", "", "Flavor name (required)")
	cloudserverCreateCmd.Flags().String("keypair-uri", "", "Keypair URI (e.g., /projects/{project-id}/providers/Aruba.Compute/keyPairs/{keypair-name})")
	cloudserverCreateCmd.Flags().StringSlice("tags", []string{}, "Tags (comma-separated)")
	cloudserverCreateCmd.Flags().String("user-data-file", "", "Path to cloud-init YAML file (will be base64 encoded)")
	cloudserverCreateCmd.Flags().String("vpc-uri", "", "VPC URI (required, e.g., /projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id})")
	cloudserverCreateCmd.Flags().StringSlice("subnet-uri", []string{}, "Subnet URI(s) (required, comma-separated)")
	cloudserverCreateCmd.Flags().StringSlice("security-group-uri", []string{}, "Security Group URI(s) (required, comma-separated)")
	cloudserverCreateCmd.Flags().String("elasticip-uri", "", "Elastic IP URI (optional)")
	cloudserverCreateCmd.Flags().String("billing-period", "Hour", "Billing period: Hour, Month, Year (optional, default: Hour)")
	cloudserverCreateCmd.MarkFlagRequired("name")
	cloudserverCreateCmd.MarkFlagRequired("region")
	cloudserverCreateCmd.MarkFlagRequired("flavor")
	cloudserverCreateCmd.MarkFlagRequired("zone")

	cloudserverGetCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	cloudserverUpdateCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverUpdateCmd.Flags().String("name", "", "New name for the cloud server")
	cloudserverUpdateCmd.Flags().StringSlice("tags", []string{}, "New tags (comma-separated)")

	cloudserverDeleteCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverDeleteCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
	cloudserverDeleteCmd.Flags().Bool("dry-run", false, "Validate resource exists without deleting")

	cloudserverListCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverListCmd.Flags().Int32("limit", 0, "Maximum number of results to return (0 = no limit)")
	cloudserverListCmd.Flags().Int32("offset", 0, "Number of results to skip")

	cloudserverPowerOnCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverPowerOffCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")

	cloudserverSetPasswordCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverSetPasswordCmd.Flags().String("password", "", "New password for the cloud server (required)")
	cloudserverSetPasswordCmd.MarkFlagRequired("password")

	cloudserverConnectCmd.Flags().String("project-id", "", "Project ID (uses context if not specified)")
	cloudserverConnectCmd.Flags().String("user", "<user>", "SSH username (required - see documentation for image-specific users)")

	// Set up auto-completion for resource IDs
	cloudserverGetCmd.ValidArgsFunction = completeCloudServerID
	cloudserverUpdateCmd.ValidArgsFunction = completeCloudServerID
	cloudserverDeleteCmd.ValidArgsFunction = completeCloudServerID
	cloudserverPowerOnCmd.ValidArgsFunction = completeCloudServerID
	cloudserverPowerOffCmd.ValidArgsFunction = completeCloudServerID
	cloudserverSetPasswordCmd.ValidArgsFunction = completeCloudServerID
	cloudserverConnectCmd.ValidArgsFunction = completeCloudServerID
}

// cloudServerRef builds the combined project+server Ref that v0.2.0 Get/Delete need.
func cloudServerRef(projectID, serverID string) aruba.Ref {
	return aruba.URI("/projects/" + projectID +
		"/providers/Aruba.Compute/cloudServers/" + serverID)
}

// csListPayload extracts the typed list for -o json/yaml; the List wrapper is not
// JSON-marshalable. Mirrors projectListPayload from the #102 playbook.
func csListPayload(l *aruba.List[*aruba.CloudServer]) any {
	if r, ok := l.Raw().(*types.Response[types.CloudServerList]); ok && r != nil {
		return r.Data
	}
	return nil
}

// Completion functions for compute resources
func completeCloudServerID(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectID, err := GetProjectID(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	client, err := GetArubaClient()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	list, err := client.FromCompute().CloudServers().List(ctx, projectRef(projectID))
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var completions []string
	if list != nil {
		for _, cs := range list.Items() {
			id := cs.ID()
			if id == "" {
				continue
			}
			if toComplete == "" || strings.HasPrefix(id, toComplete) {
				completions = append(completions, fmt.Sprintf("%s\t%s", id, cs.Name()))
			}
		}
	}

	return completions, cobra.ShellCompDirectiveNoFileComp
}

// CloudServer subcommands
var cloudserverCmd = &cobra.Command{
	Use:   "cloudserver",
	Short: "Manage cloud servers",
	Long:  `Perform CRUD operations on cloud servers in Aruba Cloud.`,
}

var cloudserverCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cloud server",
	Long: `Create a new cloud server in the specified region and VPC.

The boot disk is specified as a URI referencing a compute template (image). Network
resources (VPC, subnet, security group) must already exist; pass their URIs with
--vpc-uri, --subnet-uri, and --security-group-uri.

Billing period: Hour (default), Month, or Year.`,
	Example: `  acloud compute cloudserver create \
    --name my-server --region IT-BG --zone IT-BG-1 \
    --flavor <flavor-id> \
    --boot-disk-uri /projects/<proj-id>/providers/Aruba.Compute/templates/<template-id> \
    --vpc-uri /projects/<proj-id>/providers/Aruba.Network/vpcs/<vpc-id> \
    --subnet-uri /projects/<proj-id>/providers/Aruba.Network/subnets/<subnet-id> \
    --security-group-uri /projects/<proj-id>/providers/Aruba.Network/securityGroups/<sg-id>`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		vpcURI, _ := cmd.Flags().GetString("vpc-uri")
		subnetURIs, _ := cmd.Flags().GetStringSlice("subnet-uri")
		securityGroupURIs, _ := cmd.Flags().GetStringSlice("security-group-uri")
		elasticIPURI, _ := cmd.Flags().GetString("elasticip-uri")
		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		region, _ := cmd.Flags().GetString("region")
		zone, _ := cmd.Flags().GetString("zone")
		flavor, _ := cmd.Flags().GetString("flavor")
		bootDiskURI, _ := cmd.Flags().GetString("boot-disk-uri")
		keypairURI, _ := cmd.Flags().GetString("keypair-uri")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		billingPeriod, _ := cmd.Flags().GetString("billing-period")
		userDataFile, _ := cmd.Flags().GetString("user-data-file")

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		cs := aruba.NewCloudServer().
			IntoProject(projectRef(projectID)).
			Named(name).
			InRegion(aruba.Region(region)).
			InZone(aruba.Zone(zone)).
			OfFlavor(aruba.CloudServerFlavor(flavor)).
			WithBootVolume(aruba.URI(bootDiskURI)).
			WithVPC(aruba.URI(vpcURI))
		if len(tags) > 0 {
			cs.ReplaceTags(tags...)
		}
		for _, s := range subnetURIs {
			cs.AddSubnet(aruba.URI(s))
		}
		for _, sg := range securityGroupURIs {
			cs.AddSecurityGroup(aruba.URI(sg))
		}
		if keypairURI != "" {
			cs.WithKeyPair(aruba.URI(keypairURI))
		}
		if elasticIPURI != "" {
			cs.WithElasticIP(aruba.URI(elasticIPURI))
		}
		if billingPeriod != "" {
			cs.WithBillingPeriod(aruba.BillingPeriod(billingPeriod))
		}
		if userDataFile != "" {
			fileContent, err := os.ReadFile(userDataFile)
			if err != nil {
				return fmt.Errorf("reading user-data file: %w", err)
			}
			cs.WithUserData(base64.StdEncoding.EncodeToString(fileContent))
		}

		ctx, cancel := newCtx()
		defer cancel()
		created, err := client.FromCompute().CloudServers().Create(ctx, cs)
		if err != nil {
			return fmt.Errorf("creating cloud server: %w", apiErrFromV2(err))
		}

		r := created.Raw()
		if r != nil {
			headers := []TableColumn{
				{Header: "ID", Width: 30},
				{Header: "NAME", Width: 40},
				{Header: "FLAVOR", Width: 20},
				{Header: "CPU", Width: 10},
				{Header: "RAM(GB)", Width: 15},
				{Header: "HD(GB)", Width: 15},
				{Header: "REGION", Width: 20},
			}
			var id, csName string
			if r.Metadata.ID != nil {
				id = *r.Metadata.ID
			}
			if r.Metadata.Name != nil {
				csName = *r.Metadata.Name
			}
			flavorName := string(r.Properties.Flavor.Name)
			cpu := r.Properties.Flavor.CPU
			ram := r.Properties.Flavor.RAM
			hd := r.Properties.Flavor.HD
			regionValue := ""
			if r.Metadata.LocationResponse != nil {
				regionValue = string(r.Metadata.LocationResponse.Value)
			}
			row := []string{
				id,
				csName,
				flavorName,
				fmt.Sprintf("%d", cpu),
				fmt.Sprintf("%d", ram),
				fmt.Sprintf("%d", hd),
				regionValue,
			}
			PrintOutput(r, headers, [][]string{row})
		} else {
			fmt.Println(msgCreatedAsync("Cloud server", name))
		}
		return nil
	},
}

var cloudserverGetCmd = &cobra.Command{
	Use:   "get [cloudserver-id]",
	Short: "Get cloud server details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		got, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
		if err != nil {
			return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
		}

		r := got.Raw()
		if r != nil {
			format := resolveOutputFormat()
			if format == OutputFormatJSON || format == OutputFormatYAML {
				PrintOutput(r, nil, nil)
				return nil
			}

			fmt.Println("\nCloud Server Details:")
			fmt.Println("====================")

			if r.Metadata.ID != nil {
				fmt.Printf("ID:              %s\n", *r.Metadata.ID)
			}
			if r.Metadata.Name != nil {
				fmt.Printf("Name:            %s\n", *r.Metadata.Name)
			}
			if r.Metadata.LocationResponse != nil && r.Metadata.LocationResponse.Value != "" {
				fmt.Printf("Region:          %s\n", r.Metadata.LocationResponse.Value)
			}

			if r.Properties.Flavor.Name != "" {
				fmt.Printf("Flavor:          %s\n", r.Properties.Flavor.Name)
			}
			fmt.Printf("CPU:             %d\n", r.Properties.Flavor.CPU)
			fmt.Printf("RAM:             %d GB\n", r.Properties.Flavor.RAM)
			fmt.Printf("HD:              %d GB\n", r.Properties.Flavor.HD)

			if r.Properties.BootVolume.URI != "" {
				fmt.Printf("Boot Volume URI: %s\n", r.Properties.BootVolume.URI)
			}

			if r.Properties.KeyPair.URI != "" {
				fmt.Printf("Keypair URI:     %s\n", r.Properties.KeyPair.URI)
			}

			if r.Status.State != nil {
				fmt.Printf("Status:          %s\n", *r.Status.State)
			}

			if len(r.Metadata.Tags) > 0 {
				fmt.Printf("Tags:            %v\n", r.Metadata.Tags)
			} else {
				fmt.Printf("Tags:            []\n")
			}

			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				jsonData, _ := json.MarshalIndent(r, "", "  ")
				fmt.Println("\nFull JSON Response:")
				fmt.Println("==================")
				fmt.Println(string(jsonData))
			}
		} else {
			fmt.Println("Cloud server not found or no data returned.")
		}
		return nil
	},
}

var cloudserverUpdateCmd = &cobra.Command{
	Use:   "update [cloudserver-id]",
	Short: "Update a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		cur, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
		if err != nil {
			return fmt.Errorf("fetching current cloud server: %w", apiErrFromV2(err))
		}

		if name != "" {
			cur.Named(name)
		}
		if cmd.Flags().Changed("tags") {
			cur.ReplaceTags(tags...)
		}

		updated, err := client.FromCompute().CloudServers().Update(ctx, cur)
		if err != nil {
			return fmt.Errorf("updating cloud server: %w", apiErrFromV2(err))
		}

		r := updated.Raw()
		if r != nil {
			fmt.Printf("\n%s\n", msgUpdated("Cloud server", serverID))
			if r.Metadata.Name != nil {
				fmt.Printf("Name:    %s\n", *r.Metadata.Name)
			}
			if len(r.Metadata.Tags) > 0 {
				fmt.Printf("Tags:    %v\n", r.Metadata.Tags)
			}
		} else {
			fmt.Println(msgUpdatedAsync("Cloud server", serverID))
		}
		return nil
	},
}

var cloudserverDeleteCmd = &cobra.Command{
	Use:   "delete [cloudserver-id]",
	Short: "Delete a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		skipConfirm, _ := cmd.Flags().GetBool("yes")
		if !skipConfirm {
			ok, err := confirmDelete("cloud server", serverID)
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
			_, err = client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
			if err != nil {
				return fmt.Errorf("dry-run: cloud server not found or inaccessible: %w", apiErrFromV2(err))
			}
			fmt.Println(msgDryRun("cloud server", serverID))
			return nil
		}

		if err := client.FromCompute().CloudServers().Delete(ctx, cloudServerRef(projectID, serverID)); err != nil {
			return fmt.Errorf("deleting cloud server: %w", apiErrFromV2(err))
		}

		fmt.Println(msgDeleted("Cloud server", serverID))
		return nil
	},
}

var cloudserverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all cloud servers",
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
		list, err := client.FromCompute().CloudServers().List(ctx, projectRef(projectID), listOpts(cmd)...)
		if err != nil {
			return fmt.Errorf("listing cloud servers: %w", apiErrFromV2(err))
		}

		if list != nil && len(list.Items()) > 0 {
			headers := []TableColumn{
				{Header: "NAME", Width: 25},
				{Header: "ID", Width: 30},
				{Header: "LOCATION", Width: 15},
				{Header: "FLAVOR", Width: 15},
				{Header: "STATUS", Width: 15},
			}

			var rows [][]string
			for _, cs := range list.Items() {
				if cs.ID() == "" {
					continue
				}
				rows = append(rows, []string{
					cs.Name(),
					cs.ID(),
					string(cs.Region()),
					string(cs.Flavor()),
					cs.State(),
				})
			}

			if len(rows) == 0 {
				fmt.Println("No cloud servers found")
				return nil
			}
			PrintOutput(csListPayload(list), headers, rows)
		} else {
			fmt.Println("No cloud servers found")
		}
		return nil
	},
}

var cloudserverPowerOnCmd = &cobra.Command{
	Use:   "power-on [cloudserver-id]",
	Short: "Power on a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		cs, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
		if err != nil {
			return fmt.Errorf("powering on cloud server: %w", apiErrFromV2(err))
		}

		if err := cs.PowerOn(ctx); err != nil {
			return fmt.Errorf("powering on cloud server: %w", apiErrFromV2(err))
		}

		if cs.Raw() != nil {
			fmt.Println(msgAction("Cloud server", serverID, "powered on"))
			if cs.Raw().Metadata.Name != nil {
				fmt.Printf("Server: %s\n", *cs.Raw().Metadata.Name)
			}
			if cs.Raw().Status.State != nil {
				fmt.Printf("Status: %s\n", *cs.Raw().Status.State)
			}
		} else {
			fmt.Println(msgAction("Cloud server", serverID, "power-on initiated"))
		}
		return nil
	},
}

var cloudserverPowerOffCmd = &cobra.Command{
	Use:   "power-off [cloudserver-id]",
	Short: "Power off a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

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
		cs, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
		if err != nil {
			return fmt.Errorf("powering off cloud server: %w", apiErrFromV2(err))
		}

		if err := cs.PowerOff(ctx); err != nil {
			return fmt.Errorf("powering off cloud server: %w", apiErrFromV2(err))
		}

		if cs.Raw() != nil {
			fmt.Println(msgAction("Cloud server", serverID, "powered off"))
			if cs.Raw().Metadata.Name != nil {
				fmt.Printf("Server: %s\n", *cs.Raw().Metadata.Name)
			}
			if cs.Raw().Status.State != nil {
				fmt.Printf("Status: %s\n", *cs.Raw().Status.State)
			}
		} else {
			fmt.Println(msgAction("Cloud server", serverID, "power-off initiated"))
		}
		return nil
	},
}

var cloudserverSetPasswordCmd = &cobra.Command{
	Use:   "set-password [cloudserver-id]",
	Short: "Set password for a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		password, _ := cmd.Flags().GetString("password")
		if password == "" {
			return fmt.Errorf("--password is required")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()
		cs, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
		if err != nil {
			return fmt.Errorf("setting cloud server password: %w", apiErrFromV2(err))
		}

		// SetPassword does not re-hydrate cs; name/state are from the prior Get.
		if err := cs.SetPassword(ctx, password); err != nil {
			return fmt.Errorf("setting cloud server password: %w", apiErrFromV2(err))
		}

		fmt.Println(msgAction("Cloud server", serverID, "password set"))
		r := cs.Raw()
		if r != nil {
			if r.Metadata.Name != nil {
				fmt.Printf("Server: %s\n", *r.Metadata.Name)
			}
			if r.Status.State != nil {
				fmt.Printf("Status: %s\n", *r.Status.State)
			}
		}
		return nil
	},
}

var cloudserverConnectCmd = &cobra.Command{
	Use:   "connect [cloudserver-id]",
	Short: "Get SSH connection information for a cloud server",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverID := args[0]

		projectID, err := GetProjectID(cmd)
		if err != nil {
			return err
		}

		user, _ := cmd.Flags().GetString("user")
		if user == "" || user == "<user>" {
			fmt.Println("Error: --user is required")
			fmt.Println("\nCommon SSH users by image type:")
			fmt.Println("  - Ubuntu/Debian: ubuntu")
			fmt.Println("  - CentOS/RHEL: centos or root")
			fmt.Println("  - Other Linux: root or check image documentation")
			fmt.Println("\nFor more information, see: https://kb.arubacloud.com/cmp/en/computing/cloud-server.aspx")
			return fmt.Errorf("--user is required")
		}

		client, err := GetArubaClient()
		if err != nil {
			return fmt.Errorf("initializing client: %w", err)
		}

		ctx, cancel := newCtx()
		defer cancel()

		cs, err := client.FromCompute().CloudServers().Get(ctx, cloudServerRef(projectID, serverID))
		if err != nil {
			return fmt.Errorf("getting cloud server: %w", apiErrFromV2(err))
		}

		r := cs.Raw()
		if r == nil {
			fmt.Println("Cloud server not found or no data returned.")
			return nil
		}

		var elasticIPURI string
		for _, linkedResource := range r.Properties.LinkedResources {
			if strings.Contains(linkedResource.URI, "providers/Aruba.Network/elasticIps") {
				elasticIPURI = linkedResource.URI
				break
			}
		}

		if elasticIPURI == "" {
			fmt.Println("No Elastic IP found for this cloud server.")
			fmt.Println("The server must have an Elastic IP linked to use the connect command.")
			return nil
		}

		eip, err := client.FromNetwork().ElasticIPs().Get(ctx, aruba.URI(elasticIPURI))
		if err != nil {
			return fmt.Errorf("getting Elastic IP details: %w", apiErrFromV2(err))
		}

		addr := eip.Address()
		if addr == "" {
			fmt.Println("Elastic IP address not available.")
			return nil
		}

		fmt.Printf("Connect by running: ssh %s@%s\n", user, addr)
		return nil
	},
}
