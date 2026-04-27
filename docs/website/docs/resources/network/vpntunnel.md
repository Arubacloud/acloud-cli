# VPN Tunnel

VPN Tunnels in Aruba Cloud provide secure, encrypted site-to-site connections between your VPC and remote networks (such as on-premises data centers or other clouds).

## Commands

### List VPN Tunnels

```bash
acloud network vpntunnel list [flags]
```

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)
- `--limit int` - Maximum number of results to return
- `--offset int` - Number of results to skip

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

---

### Get VPN Tunnel Details

```bash
acloud network vpntunnel get <vpn-tunnel-id> [flags]
```

**Flags:**
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpntunnel get 1234567890abcdef
```

---

### Create VPN Tunnel

```bash
acloud network vpntunnel create [flags]
```

**Required Flags:**
- `--name string` - Name for the VPN tunnel
- `--region string` - Region code (e.g., `ITBG-Bergamo`)
- `--peer-ip string` - Peer client public IPv4 address
- `--vpc-uri string` - VPC URI (`/projects/{project-id}/providers/Aruba.Network/vpcs/{vpc-id}`)
- `--elastic-ip-uri string` - Elastic IP URI (`/projects/{project-id}/providers/Aruba.Network/elasticIps/{eip-id}`)
- `--subnet-cidr string` or `--subnet-name string` - VPN subnet CIDR or name (one is required)

**Optional Flags:**
- `--tags strings` - Tags (comma-separated)
- `--vpn-type string` - VPN type (default: `Site-To-Site`)
- `--protocol string` - VPN protocol (default: `ikev2`)
- `--billing-period string` - Billing period: `Hour`, `Month`, `Year` (default: `Hour`)

**IKE Group Flags:**
- `--ike-lifetime int32` - IKE lifetime in seconds (0–86400)
- `--ike-encryption string` - IKE encryption algorithm (see [Encryption algorithms](#encryption-algorithms))
- `--ike-hash string` - IKE hash algorithm (see [Hash algorithms](#hash-algorithms))
- `--ike-dh-group string` - IKE DH group number: `1`, `2`, `5`, or `14`–`32`
- `--ike-dpd-action string` - IKE DPD action: `trap`, `clear`, `restart`
- `--ike-dpd-interval int32` - IKE DPD interval in seconds (2–86400)
- `--ike-dpd-timeout int32` - IKE DPD timeout in seconds (2–86400)

**Authentication (PSK) Flags:**
- `--psk string` - Pre-shared key secret (max 50 chars)
- `--psk-cloud-site string` - PSK ID for the Aruba (cloud) side (3–100 chars, alphanumeric, `-` and `.`)
- `--psk-onprem-site string` - PSK ID for the customer (on-prem) side (3–100 chars, alphanumeric, `-` and `.`)

**ESP Group Flags:**
- `--esp-lifetime int32` - ESP lifetime in seconds (30–86400)
- `--esp-encryption string` - ESP encryption algorithm (default: `aes256`; see [Encryption algorithms](#encryption-algorithms))
- `--esp-hash string` - ESP hash algorithm (see [Hash algorithms](#hash-algorithms))
- `--esp-pfs string` - ESP PFS group (see [PFS groups](#pfs-groups))

**Example:**
```bash
acloud network vpntunnel create \
  --name my-tunnel \
  --region ITBG-Bergamo \
  --peer-ip 203.0.113.1 \
  --vpc-uri /projects/<proj-id>/providers/Aruba.Network/vpcs/<vpc-id> \
  --subnet-cidr 10.241.0.0/24 \
  --subnet-name my-vpn-subnet \
  --elastic-ip-uri /projects/<proj-id>/providers/Aruba.Network/elasticIps/<eip-id> \
  --ike-lifetime 3600 \
  --ike-encryption aes256 \
  --ike-hash sha256 \
  --ike-dh-group 14 \
  --ike-dpd-action restart \
  --ike-dpd-interval 10 \
  --ike-dpd-timeout 30 \
  --psk my-pre-shared-key \
  --psk-cloud-site psk-aruba-side \
  --psk-onprem-site psk-customer-side \
  --esp-lifetime 1800 \
  --esp-encryption aes256 \
  --esp-hash sha256 \
  --esp-pfs enable \
  --billing-period Hour
```

---

### Update VPN Tunnel

Only `--name` and `--tags` are updatable. All other fields (IKE, ESP, subnet) are immutable after creation.

```bash
acloud network vpntunnel update <vpn-tunnel-id> [flags]
```

**Flags:**
- `--name string` - New name for the VPN tunnel
- `--tags strings` - New tags (comma-separated)
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpntunnel update 1234567890abcdef --name new-name --tags env=prod
```

---

### Delete VPN Tunnel

```bash
acloud network vpntunnel delete <vpn-tunnel-id> [flags]
```

**Flags:**
- `-y, --yes` - Skip confirmation prompt
- `--dry-run` - Validate existence without deleting
- `--project-id string` - Project ID (uses context if not specified)

**Example:**
```bash
acloud network vpntunnel delete 1234567890abcdef --yes
```

---

## Valid Enum Values

The CLI validates enum fields client-side before sending the request, so typos return a clear error immediately.

### Encryption Algorithms

Valid for both `--ike-encryption` and `--esp-encryption`:

| Value | Description |
|---|---|
| `aes128` | 128-bit AES-CBC |
| `aes192` | 192-bit AES-CBC |
| `aes256` | 256-bit AES-CBC |
| `aes128ctr` | 128-bit AES-COUNTER |
| `aes192ctr` | 192-bit AES-COUNTER |
| `aes256ctr` | 256-bit AES-COUNTER |
| `aes128ccm64` | 128-bit AES-CCM with 64-bit ICV |
| `aes192ccm64` | 192-bit AES-CCM with 64-bit ICV |
| `aes256ccm64` | 256-bit AES-CCM with 64-bit ICV |
| `aes128ccm96` | 128-bit AES-CCM with 96-bit ICV |
| `aes192ccm96` | 192-bit AES-CCM with 96-bit ICV |
| `aes256ccm96` | 256-bit AES-CCM with 96-bit ICV |
| `aes128ccm128` | 128-bit AES-CCM with 128-bit ICV |
| `aes192ccm128` | 192-bit AES-CCM with 128-bit ICV |
| `aes256ccm128` | 256-bit AES-CCM with 128-bit ICV |
| `aes128gcm64` | 128-bit AES-GCM with 64-bit ICV |
| `aes192gcm64` | 192-bit AES-GCM with 64-bit ICV |
| `aes256gcm64` | 256-bit AES-GCM with 64-bit ICV |
| `aes128gcm96` | 128-bit AES-GCM with 96-bit ICV |
| `aes192gcm96` | 192-bit AES-GCM with 96-bit ICV |
| `aes256gcm96` | 256-bit AES-GCM with 96-bit ICV |
| `aes128gcm128` | 128-bit AES-GCM with 128-bit ICV |
| `aes192gcm128` | 192-bit AES-GCM with 128-bit ICV |
| `aes256gcm128` | 256-bit AES-GCM with 128-bit ICV |
| `aes128gmac` | Null encryption with 128-bit AES-GMAC |
| `aes192gmac` | Null encryption with 192-bit AES-GMAC |
| `aes256gmac` | Null encryption with 256-bit AES-GMAC |
| `3des` | 168-bit 3DES-EDE-CBC |
| `blowfish128` | 128-bit Blowfish-CBC |
| `blowfish192` | 192-bit Blowfish-CBC |
| `blowfish256` | 256-bit Blowfish-CBC |
| `camellia128` | 128-bit Camellia-CBC |
| `camellia192` | 192-bit Camellia-CBC |
| `camellia256` | 256-bit Camellia-CBC |
| `camellia128ctr` | 128-bit Camellia-COUNTER |
| `camellia192ctr` | 192-bit Camellia-COUNTER |
| `camellia256ctr` | 256-bit Camellia-COUNTER |
| `camellia128ccm64` | 128-bit Camellia-CCM with 64-bit ICV |
| `camellia192ccm64` | 192-bit Camellia-CCM with 64-bit ICV |
| `camellia256ccm64` | 256-bit Camellia-CCM with 64-bit ICV |
| `camellia128ccm96` | 128-bit Camellia-CCM with 96-bit ICV |
| `camellia192ccm96` | 192-bit Camellia-CCM with 96-bit ICV |
| `camellia256ccm96` | 256-bit Camellia-CCM with 96-bit ICV |
| `camellia128ccm128` | 128-bit Camellia-CCM with 128-bit ICV |
| `camellia192ccm128` | 192-bit Camellia-CCM with 128-bit ICV |
| `camellia256ccm128` | 256-bit Camellia-CCM with 128-bit ICV |
| `serpent128` | 128-bit Serpent-CBC |
| `serpent192` | 192-bit Serpent-CBC |
| `serpent256` | 256-bit Serpent-CBC |
| `twofish128` | 128-bit Twofish-CBC |
| `twofish192` | 192-bit Twofish-CBC |
| `twofish256` | 256-bit Twofish-CBC |
| `cast128` | 128-bit CAST-CBC |
| `chacha20poly1305` | 256-bit ChaCha20/Poly1305 with 128-bit ICV |

### Hash Algorithms

Valid for both `--ike-hash` and `--esp-hash`:

| Value | Description |
|---|---|
| `md5` | MD5 HMAC |
| `md5_128` | MD5-128 HMAC |
| `sha1` | SHA1 HMAC |
| `sha1_160` | SHA1-160 HMAC |
| `sha256` | SHA2-256-128 HMAC |
| `sha256_96` | SHA2-256-96 HMAC |
| `sha384` | SHA2-384-192 HMAC |
| `sha512` | SHA2-512-256 HMAC |
| `aesxcbc` | AES XCBC |
| `aescmac` | AES CMAC |
| `aes128gmac` | 128-bit AES-GMAC |
| `aes192gmac` | 192-bit AES-GMAC |
| `aes256gmac` | 256-bit AES-GMAC |

### DH Groups

Valid for `--ike-dh-group`:

| Value | Description |
|---|---|
| `1` | Diffie-Hellman group 1 (modp768) |
| `2` | Diffie-Hellman group 2 (modp1024) |
| `5` | Diffie-Hellman group 5 (modp1536) |
| `14` | Diffie-Hellman group 14 (modp2048) |
| `15` | Diffie-Hellman group 15 (modp3072) |
| `16` | Diffie-Hellman group 16 (modp4096) |
| `17` | Diffie-Hellman group 17 (modp6144) |
| `18` | Diffie-Hellman group 18 (modp8192) |
| `19` | Diffie-Hellman group 19 (ecp256) |
| `20` | Diffie-Hellman group 20 (ecp384) |
| `21` | Diffie-Hellman group 21 (ecp521) |
| `22` | Diffie-Hellman group 22 (modp1024s160) |
| `23` | Diffie-Hellman group 23 (modp2048s224) |
| `24` | Diffie-Hellman group 24 (modp2048s256) |
| `25` | Diffie-Hellman group 25 (ecp192) |
| `26` | Diffie-Hellman group 26 (ecp224) |
| `27` | Diffie-Hellman group 27 (ecp224bp) |
| `28` | Diffie-Hellman group 28 (ecp256bp) |
| `29` | Diffie-Hellman group 29 (ecp384bp) |
| `30` | Diffie-Hellman group 30 (ecp512bp) |
| `31` | Diffie-Hellman group 31 (curve25519) |
| `32` | Diffie-Hellman group 32 (curve448) |

### DPD Actions

Valid for `--ike-dpd-action`:

| Value | Description |
|---|---|
| `trap` | Hold the SA and re-initiate on traffic |
| `clear` | Close the SA immediately |
| `restart` | Re-negotiate the SA immediately |

### PFS Groups

Valid for `--esp-pfs`:

| Value | Description |
|---|---|
| `enable` | Inherit DH group from the IKE group |
| `dh-group1` | Use Diffie-Hellman group 1 (modp768) |
| `dh-group2` | Use Diffie-Hellman group 2 (modp1024) |
| `dh-group5` | Use Diffie-Hellman group 5 (modp1536) |
| `dh-group14` | Use Diffie-Hellman group 14 (modp2048) |
| `dh-group15` | Use Diffie-Hellman group 15 (modp3072) |
| `dh-group16` | Use Diffie-Hellman group 16 (modp4096) |
| `dh-group17` | Use Diffie-Hellman group 17 (modp6144) |
| `dh-group18` | Use Diffie-Hellman group 18 (modp8192) |
| `dh-group19` | Use Diffie-Hellman group 19 (ecp256) |
| `dh-group20` | Use Diffie-Hellman group 20 (ecp384) |
| `dh-group21` | Use Diffie-Hellman group 21 (ecp521) |
| `dh-group22` | Use Diffie-Hellman group 22 (modp1024s160) |
| `dh-group23` | Use Diffie-Hellman group 23 (modp2048s224) |
| `dh-group24` | Use Diffie-Hellman group 24 (modp2048s256) |
| `dh-group25` | Use Diffie-Hellman group 25 (ecp192) |
| `dh-group26` | Use Diffie-Hellman group 26 (ecp224) |
| `dh-group27` | Use Diffie-Hellman group 27 (ecp224bp) |
| `dh-group28` | Use Diffie-Hellman group 28 (ecp256bp) |
| `dh-group29` | Use Diffie-Hellman group 29 (ecp384bp) |
| `dh-group30` | Use Diffie-Hellman group 30 (ecp512bp) |
| `dh-group31` | Use Diffie-Hellman group 31 (curve25519) |
| `dh-group32` | Use Diffie-Hellman group 32 (curve448) |
| `disable` | Disable PFS |

---

## Shell Auto-completion

The `get`, `update`, and `delete` commands support auto-completion for VPN tunnel IDs.

## Related Commands

- [VPN Route](vpnroute.md)
- [VPC](vpc.md)
- [Elastic IP](elasticip.md)
