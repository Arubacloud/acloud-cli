package cmd

import "github.com/Arubacloud/sdk-go/pkg/aruba"

// Cross-resource valid-value slices used by Args.Validate() methods.
//
// Rules:
//   - Populated from sdk-go v1.0.0 aruba.* constants in pkg/aruba/aliases.go.
//   - For enums where the SDK exposes no fixed set (e.g. aruba.KubernetesVersion),
//     Validate() checks != "" only — never slices.Contains. Annotate inline per resource.
//   - Resource-local enum slices (e.g. the seven vpn*Algorithms in network.vpntunnel.go)
//     stay in their resource file.

// validRegions is the set of Aruba Cloud region codes recognised by the SDK.
// Currently only one datacenter is available.
var validRegions = []aruba.Region{
	aruba.RegionITBGBergamo, // "ITBG-Bergamo"
}

// validZones is the set of availability zones within the ITBG-Bergamo region.
var validZones = []aruba.Zone{
	aruba.ZoneITBG1, // "ITBG-1"
	aruba.ZoneITBG2, // "ITBG-2"
	aruba.ZoneITBG3, // "ITBG-3"
}

// validBillingPeriods is the set of billing cadences accepted by the API.
var validBillingPeriods = []aruba.BillingPeriod{
	aruba.BillingPeriodHour,  // "Hour"
	aruba.BillingPeriodMonth, // "Month"
	aruba.BillingPeriodYear,  // "Year"
}

// validCloudServerFlavors is the set of Cloud Server SKUs recognised by the SDK.
var validCloudServerFlavors = []aruba.CloudServerFlavor{
	aruba.CloudServerFlavorCSO1A2,   //  1 vCPU,  2 GB RAM
	aruba.CloudServerFlavorCSO1A4,   //  1 vCPU,  4 GB RAM
	aruba.CloudServerFlavorCSO2A4,   //  2 vCPU,  4 GB RAM
	aruba.CloudServerFlavorCSO2A8,   //  2 vCPU,  8 GB RAM
	aruba.CloudServerFlavorCSO4A8,   //  4 vCPU,  8 GB RAM
	aruba.CloudServerFlavorCSO4A16,  //  4 vCPU, 16 GB RAM
	aruba.CloudServerFlavorCSO8A16,  //  8 vCPU, 16 GB RAM
	aruba.CloudServerFlavorCSO8A32,  //  8 vCPU, 32 GB RAM
	aruba.CloudServerFlavorCSO12A24, // 12 vCPU, 24 GB RAM
	aruba.CloudServerFlavorCSO16A32, // 16 vCPU, 32 GB RAM
	aruba.CloudServerFlavorCSO16A64, // 16 vCPU, 64 GB RAM
	aruba.CloudServerFlavorCSO24A48, // 24 vCPU, 48 GB RAM
	aruba.CloudServerFlavorCSO32A64, // 32 vCPU, 64 GB RAM
}

// validStorageBackupTypes is the set of backup capture modes.
var validStorageBackupTypes = []aruba.StorageBackupType{
	aruba.StorageBackupTypeFull,        // "Full"
	aruba.StorageBackupTypeIncremental, // "Incremental"
}

// validJobTypes is the set of schedule job scheduling modes.
var validJobTypes = []aruba.JobType{
	aruba.JobTypeOneShot,   // "OneShot"
	aruba.JobTypeRecurring, // "Recurring"
}

// validNodePoolInstances is the set of KaaS node pool instance types.
var validNodePoolInstances = []aruba.NodePoolInstance{
	aruba.NodePoolInstanceK1A2,   //  1 vCPU,  2 GB RAM (balanced)
	aruba.NodePoolInstanceK1A4R,  //  1 vCPU,  4 GB RAM (RAM-optimized)
	aruba.NodePoolInstanceK2A4,   //  2 vCPU,  4 GB RAM (balanced)
	aruba.NodePoolInstanceK2A8R,  //  2 vCPU,  8 GB RAM (RAM-optimized)
	aruba.NodePoolInstanceK4A8,   //  4 vCPU,  8 GB RAM (balanced)
	aruba.NodePoolInstanceK4A16R, //  4 vCPU, 16 GB RAM (RAM-optimized)
	aruba.NodePoolInstanceK8A16,  //  8 vCPU, 16 GB RAM (balanced)
	aruba.NodePoolInstanceK8A32R, //  8 vCPU, 32 GB RAM (RAM-optimized)
	aruba.NodePoolInstanceK12A24, // 12 vCPU, 24 GB RAM (balanced)
	aruba.NodePoolInstanceK16A32, // 16 vCPU, 32 GB RAM (balanced)
	aruba.NodePoolInstanceK24A48, // 24 vCPU, 48 GB RAM (balanced)
	aruba.NodePoolInstanceK32A64, // 32 vCPU, 64 GB RAM (balanced)
}

// validContainerRegistrySizeFlavors is the set of container registry concurrent-user tiers.
var validContainerRegistrySizeFlavors = []aruba.ContainerRegistrySizeFlavor{
	aruba.ContainerRegistrySizeFlavorSmall,    // "Small"
	aruba.ContainerRegistrySizeFlavorMedium,   // "Medium"
	aruba.ContainerRegistrySizeFlavorHighPerf, // "HighPerf"
}
