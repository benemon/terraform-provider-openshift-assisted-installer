package models

// ValidationInfo represents validation information from the API
type ValidationInfo struct {
	ID              string `json:"id"`
	Status          string `json:"status"` // "success", "failure", "pending", "disabled"
	Message         string `json:"message"`
	ValidationID    string `json:"validation_id,omitempty"`
	ValidationName  string `json:"validation_name,omitempty"`
	ValidationGroup string `json:"validation_group,omitempty"`
}

// ClusterValidationResponse represents cluster validation response
type ClusterValidationResponse struct {
	ValidationsInfo map[string][]ValidationInfo `json:"validations_info"`
}

// HostValidationResponse represents host validation response  
type HostValidationResponse struct {
	ID              string                      `json:"id"`
	ValidationsInfo map[string][]ValidationInfo `json:"validations_info"`
}

// HostsValidationResponse represents response for all hosts validations
type HostsValidationResponse struct {
	Hosts []HostValidationResponse `json:"hosts,omitempty"`
}

// ValidationStatus represents the possible validation statuses
type ValidationStatus string

const (
	ValidationStatusSuccess  ValidationStatus = "success"
	ValidationStatusFailure  ValidationStatus = "failure"
	ValidationStatusPending  ValidationStatus = "pending"
	ValidationStatusDisabled ValidationStatus = "disabled"
)

// ValidationType represents blocking vs non-blocking validation types
type ValidationType string

const (
	ValidationTypeBlocking    ValidationType = "blocking"
	ValidationTypeNonBlocking ValidationType = "non-blocking"
)

// Common Host Validation IDs as documented in Red Hat docs
const (
	// Connection and inventory validations
	HostValidationConnected    = "connected"
	HostValidationHasInventory = "has-inventory"
	
	// Resource requirement validations
	HostValidationHasMinCPUCores    = "has-min-cpu-cores"
	HostValidationHasMinMemory      = "has-min-memory"
	HostValidationHasMinValidDisks  = "has-min-valid-disks"
	HostValidationHasCPUCoresForRole = "has-cpu-cores-for-role"
	HostValidationHasMemoryForRole   = "has-memory-for-role"
	
	// Network validations
	HostValidationHasDefaultRoute           = "has-default-route"
	HostValidationAPIDomainNameResolved     = "api-domain-name-resolved-correctly"
	HostValidationAPIIntDomainNameResolved  = "api-int-domain-name-resolved-correctly"  
	HostValidationAppsDomainNameResolved    = "apps-domain-name-resolved-correctly"
	HostValidationNonOverlappingSubnets     = "non-overlapping-subnets"
	HostValidationBelongsToMachineCIDR      = "belongs-to-machine-cidr"
	
	// Host identity validations
	HostValidationHostnameUnique = "hostname-unique"
	HostValidationHostnameValid  = "hostname-valid"
	
	// Operator requirement validations
	HostValidationLSORequirements = "lso-requirements-satisfied"
	HostValidationODFRequirements = "odf-requirements-satisfied"
	HostValidationCNVRequirements = "cnv-requirements-satisfied"
	HostValidationLVMRequirements = "lvm-requirements-satisfied"
	
	// Disk and storage validations
	HostValidationSufficientInstallationDiskSpeed = "sufficient-installation-diskspeed"
	HostValidationNoSkipInstallationDisk         = "no-skip-installation-disk"
	HostValidationNoSkipMissingDisk              = "no-skip-missing-disk"
	HostValidationDiskEncryptionRequirements     = "disk-encryption-requirements-satisfied"
	
	// Platform and compatibility validations
	HostValidationCompatibleWithClusterPlatform = "compatible-with-cluster-platform"
	HostValidationValidPlatformNetworkSettings   = "valid-platform-network-settings"
	HostValidationCompatibleAgent               = "compatible-agent"
	
	// Network performance validations
	HostValidationSufficientNetworkLatency = "sufficient-network-latency-requirement-for-role"
	HostValidationSufficientPacketLoss     = "sufficient-packet-loss-requirement-for-role"
	
	// Time and container validations
	HostValidationNTPSynced                  = "ntp-synced"
	HostValidationContainerImagesAvailable   = "container-images-available"
	
	// DNS validations
	HostValidationDNSWildcardNotConfigured = "dns-wildcard-not-configured"
	
	// Group and cluster membership
	HostValidationBelongsToMajorityGroup = "belongs-to-majority-group"
	
	// Platform-specific validations
	HostValidationVSphereDiskUUIDEnabled = "vsphere-disk-uuid-enabled"
	
	// Installation media validation
	HostValidationMediaConnected = "media-connected"
	
	// Network MTU validation
	HostValidationMTUValid = "mtu-valid"
	
	// Ignition validation (Day 2)
	HostValidationIgnitionDownloadable = "ignition-downloadable"
)

// Common Cluster Validation IDs as documented in Red Hat docs
const (
	// Network definition validations
	ClusterValidationMachineCIDRDefined = "machine-cidr-defined"
	ClusterValidationClusterCIDRDefined = "cluster-cidr-defined"
	ClusterValidationServiceCIDRDefined = "service-cidr-defined"
	
	// Network overlap and validity validations
	ClusterValidationNoCIDRsOverlapping         = "no-cidrs-overlapping"
	ClusterValidationNetworksSameAddressFamilies = "networks-same-address-families"
	ClusterValidationNetworkPrefixValid         = "network-prefix-valid"
	ClusterValidationNetworkTypeValid           = "network-type-valid"
	
	// VIP validations
	ClusterValidationAPIVIPsDefined            = "api-vips-defined"
	ClusterValidationAPIVIPsValid              = "api-vips-valid"
	ClusterValidationIngressVIPsDefined        = "ingress-vips-defined"
	ClusterValidationIngressVIPsValid          = "ingress-vips-valid"
	ClusterValidationMachineCIDREqualsCalculated = "machine-cidr-equals-to-calculated-cidr"
	
	// Host readiness validations
	ClusterValidationAllHostsReadyToInstall = "all-hosts-are-ready-to-install"
	ClusterValidationSufficientMastersCount = "sufficient-masters-count"
	
	// Cluster configuration validations
	ClusterValidationDNSDomainDefined = "dns-domain-defined"
	ClusterValidationPullSecretSet    = "pull-secret-set"
	ClusterValidationNTPServerConfigured = "ntp-server-configured"
	
	// Operator requirement validations
	ClusterValidationLSORequirements = "lso-requirements-satisfied"
	ClusterValidationODFRequirements = "odf-requirements-satisfied"
	ClusterValidationCNVRequirements = "cnv-requirements-satisfied"
	ClusterValidationLVMRequirements = "lvm-requirements-satisfied"
)

// ValidationCategory represents different categories of validations
type ValidationCategory string

const (
	ValidationCategoryNetwork   ValidationCategory = "network"
	ValidationCategoryHost      ValidationCategory = "hardware"
	ValidationCategoryOperator  ValidationCategory = "operators"
	ValidationCategoryCluster   ValidationCategory = "cluster"
	ValidationCategoryPlatform  ValidationCategory = "platform"
	ValidationCategoryStorage   ValidationCategory = "storage"
)

// GetValidationCategory returns the category for a given validation ID
func GetValidationCategory(validationID string) ValidationCategory {
	switch validationID {
	case HostValidationHasDefaultRoute, HostValidationAPIDomainNameResolved,
		 HostValidationAPIIntDomainNameResolved, HostValidationAppsDomainNameResolved,
		 HostValidationNonOverlappingSubnets, HostValidationBelongsToMachineCIDR,
		 HostValidationSufficientNetworkLatency, HostValidationSufficientPacketLoss,
		 HostValidationMTUValid, ClusterValidationMachineCIDRDefined,
		 ClusterValidationClusterCIDRDefined, ClusterValidationServiceCIDRDefined,
		 ClusterValidationNoCIDRsOverlapping, ClusterValidationNetworksSameAddressFamilies,
		 ClusterValidationNetworkPrefixValid, ClusterValidationAPIVIPsDefined,
		 ClusterValidationAPIVIPsValid, ClusterValidationIngressVIPsDefined,
		 ClusterValidationIngressVIPsValid, ClusterValidationNetworkTypeValid:
		return ValidationCategoryNetwork
		
	case HostValidationHasMinCPUCores, HostValidationHasMinMemory,
		 HostValidationHasMinValidDisks, HostValidationHasCPUCoresForRole,
		 HostValidationHasMemoryForRole, HostValidationConnected,
		 HostValidationHasInventory:
		return ValidationCategoryHost
		
	case HostValidationLSORequirements, HostValidationODFRequirements,
		 HostValidationCNVRequirements, HostValidationLVMRequirements:
		return ValidationCategoryOperator
		
	case HostValidationSufficientInstallationDiskSpeed, HostValidationNoSkipInstallationDisk,
		 HostValidationNoSkipMissingDisk, HostValidationDiskEncryptionRequirements:
		return ValidationCategoryStorage
		
	case HostValidationCompatibleWithClusterPlatform, HostValidationValidPlatformNetworkSettings,
		 HostValidationVSphereDiskUUIDEnabled, HostValidationCompatibleAgent:
		return ValidationCategoryPlatform
		
	default:
		return ValidationCategoryCluster
	}
}

// IsBlockingValidation returns true if the validation is typically blocking
func IsBlockingValidation(validationID string) bool {
	blockingValidations := map[string]bool{
		// Host blocking validations
		HostValidationHasCPUCoresForRole:              true,
		HostValidationHasMemoryForRole:                true,
		HostValidationIgnitionDownloadable:            true,
		HostValidationBelongsToMajorityGroup:          true,
		HostValidationValidPlatformNetworkSettings:    true,
		HostValidationSufficientInstallationDiskSpeed: true,
		HostValidationSufficientNetworkLatency:        true,
		HostValidationSufficientPacketLoss:            true,
		HostValidationHasDefaultRoute:                 true,
		HostValidationAPIDomainNameResolved:           true,
		HostValidationAPIIntDomainNameResolved:        true,
		HostValidationAppsDomainNameResolved:          true,
		HostValidationDNSWildcardNotConfigured:        true,
		HostValidationNonOverlappingSubnets:           true,
		HostValidationHostnameUnique:                  true,
		HostValidationHostnameValid:                   true,
		HostValidationBelongsToMachineCIDR:            true,
		HostValidationLSORequirements:                 true,
		HostValidationODFRequirements:                 true,
		HostValidationCNVRequirements:                 true,
		HostValidationLVMRequirements:                 true,
		HostValidationCompatibleAgent:                 true,
		HostValidationNoSkipInstallationDisk:          true,
		HostValidationNoSkipMissingDisk:               true,
		HostValidationMediaConnected:                  true,
		
		// Cluster blocking validations
		ClusterValidationNoCIDRsOverlapping:         true,
		ClusterValidationNetworksSameAddressFamilies: true,
		ClusterValidationNetworkPrefixValid:         true,
		ClusterValidationMachineCIDREqualsCalculated: true,
		ClusterValidationAPIVIPsValid:               true,
		ClusterValidationIngressVIPsDefined:         true,
		ClusterValidationAllHostsReadyToInstall:     true,
		ClusterValidationSufficientMastersCount:     true,
		ClusterValidationNTPServerConfigured:        true,
		ClusterValidationNetworkTypeValid:           true,
	}
	
	return blockingValidations[validationID]
}