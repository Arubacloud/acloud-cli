# VPN Tunnel

VPN Tunnels in Aruba Cloud provide secure, encrypted connections between your VPC and remote networks (such as on-premises data centers or other clouds). You can manage site-to-site and client VPN tunnels, configure protocols, and control tunnel lifecycle.

## Commands

### List VPN Tunnels
List all VPN tunnels in your project.

```bash
acloud network vpntunnel list [flags]
```

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpntunnel list
acloud network vpntunnel list --project-id 68398923fb2cb026400d4d31
```

**Output:**
```
NAME         ID                        REGION        TYPE         STATUS
vpn-prod     1234567890abcdef          ITBG-Bergamo  Site-To-Site Active
```

### Get VPN Tunnel Details
Get detailed information about a specific VPN tunnel.

```bash
acloud network vpntunnel get <vpn-tunnel-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel (supports auto-completion)

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpntunnel get 1234567890abcdef
```

**Output:**
```
VPN Tunnel Details:
===================
ID:              1234567890abcdef
Name:            vpn-prod
Region:          ITBG-Bergamo
Type:            Site-To-Site
Status:          Active
Peer IP:         203.0.113.1
VPC:             /.../vpcs/689307f4745108d3c6343b5a
Subnet CIDR:     10.0.1.0/24
Creation Date:   06-08-2025 07:44:52
Tags:            [production vpn]
```

### Create VPN Tunnel
Create a new VPN tunnel.

```bash
acloud network vpntunnel create [flags]
```

**Required Flags:**
- `--name string` - Name for the VPN tunnel
- `--region string` - Region code (e.g., ITBG-Bergamo)
- `--peer-ip string` - Peer client public IP address
- `--vpc-uri string` - VPC URI
- `--subnet-cidr string` - Subnet CIDR (e.g., 10.0.1.0/24)

**Optional Flags:**
- `--tags strings` - Tags for the VPN tunnel (comma-separated)
- `--vpn-type string` - VPN type (default: Site-To-Site)
- `--protocol string` - VPN protocol (default: ikev2)
- `--billing-period string` - Billing period: Hour, Month, Year

**Example:**
```bash
acloud network vpntunnel create --name vpn-prod --region ITBG-Bergamo --peer-ip 203.0.113.1 --vpc-uri /projects/.../vpcs/689307f4745108d3c6343b5a --subnet-cidr 10.0.1.0/24
```

**Output:**
```
VPN Tunnel created successfully!
ID:      1234567890abcdef
Name:    vpn-prod
Region:  ITBG-Bergamo
```

### Update VPN Tunnel
Update an existing VPN tunnel's name or tags.

```bash
acloud network vpntunnel update <vpn-tunnel-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel

**Flags:**
- `--name string` - New name for the VPN tunnel
- `--tags strings` - New tags (comma-separated)

**Example:**
```bash
acloud network vpntunnel update 1234567890abcdef --name "new-vpn-name"
```

**Output:**
```
VPN Tunnel updated successfully!
ID:      1234567890abcdef
Name:    new-vpn-name
```

### Delete VPN Tunnel
Delete a VPN tunnel.

```bash
acloud network vpntunnel delete <vpn-tunnel-id> [flags]
```

**Arguments:**
- `vpn-tunnel-id` - The ID of the VPN tunnel

**Flags:**
- `-y, --yes` - Skip confirmation prompt

**Example:**
```bash
acloud network vpntunnel delete 1234567890abcdef --yes
```

**Output:**
```
VPN Tunnel 1234567890abcdef deleted successfully!
```

## Shell Auto-completion

The VPN Tunnel commands support auto-completion for VPN tunnel IDs.

## Best Practices
- Use descriptive names and tags for VPN tunnels.
- Regularly review tunnel status and configuration.

## Troubleshooting
- Ensure the VPC and subnet are **Active** before creating a tunnel.
- Check peer IP and CIDR for correctness.

## Related Commands
- [VPN Tunnel Route](vpnroute.md)
- [VPC](vpc.md)
