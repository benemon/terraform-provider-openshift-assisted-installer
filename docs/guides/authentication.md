---
page_title: "Authentication and Configuration"
subcategory: "Configuration"
---

# Authentication and Configuration

This guide covers authentication methods and provider configuration options for the OpenShift Assisted Installer provider.

## Authentication Methods

The provider supports authentication using offline tokens from the Red Hat Hybrid Cloud Console.

### Offline Token Authentication

Offline tokens provide long-term authentication that automatically refreshes access tokens as needed.

#### Obtaining an Offline Token

1. Navigate to [console.redhat.com](https://console.redhat.com)
2. Click on your profile menu (top right)
3. Select "API Tokens"
4. Click "Load token" for the offline token
5. Copy the token value

~> **Note**: Offline tokens are long-lived but do expire. Monitor your token expiry and refresh as needed.

#### Using Offline Tokens

**Method 1: Provider Configuration**
```hcl
provider "oai" {
  offline_token = var.offline_token
}

variable "offline_token" {
  description = "Red Hat offline token"
  type        = string
  sensitive   = true
}
```

**Method 2: Environment Variable**
```bash
export OFFLINE_TOKEN="your-offline-token-here"
```

```hcl
provider "oai" {
  # Token automatically read from OFFLINE_TOKEN environment variable
}
```

**Method 3: Terraform Variables File**
```hcl
# terraform.tfvars (add to .gitignore!)
offline_token = "your-offline-token-here"
```

```hcl
provider "oai" {
  offline_token = var.offline_token
}

variable "offline_token" {
  type      = string
  sensitive = true
}
```

## Provider Configuration

### Complete Configuration Example

```hcl
provider "oai" {
  # Authentication
  offline_token = var.offline_token
  
  # API Configuration
  endpoint = "https://api.openshift.com/api/assisted-install"
  timeout  = "60s"
}
```

### Configuration Options

#### endpoint
- **Type**: String
- **Default**: `https://api.openshift.com/api/assisted-install`
- **Description**: The OpenShift Assisted Service API endpoint
- **Usage**: Typically only changed for testing or private cloud deployments

#### offline_token
- **Type**: String (Sensitive)
- **Required**: Yes (unless set via environment variable)
- **Description**: Offline token for Red Hat authentication
- **Environment Variable**: `OFFLINE_TOKEN`

#### timeout
- **Type**: String
- **Default**: `30s`
- **Description**: HTTP request timeout for API calls
- **Format**: Duration string (e.g., "30s", "5m", "1h")

## Security Best Practices

### Token Storage

**✅ Recommended Approaches:**

1. **Environment Variables (Local Development)**
   ```bash
   export OFFLINE_TOKEN="your-token"
   terraform apply
   ```

2. **CI/CD Secret Management**
   ```yaml
   # GitHub Actions
   - name: Terraform Apply
     env:
       OFFLINE_TOKEN: ${{ secrets.OFFLINE_TOKEN }}
     run: terraform apply -auto-approve
   ```

3. **Terraform Cloud/Enterprise**
   - Store as sensitive workspace variable
   - Mark as "Environment Variable" type
   - Set variable name as `OFFLINE_TOKEN`

4. **External Secret Management**
   ```hcl
   data "vault_kv_secret_v2" "tokens" {
     mount = "secret"
     name  = "openshift/tokens"
   }
   
   provider "oai" {
     offline_token = data.vault_kv_secret_v2.tokens.data.offline_token
   }
   ```

**❌ Avoid:**
- Hardcoding tokens in configuration files
- Committing tokens to version control
- Sharing tokens in plain text logs or outputs

### Network Security

For enhanced security in enterprise environments:

```hcl
provider "oai" {
  endpoint = "https://your-private-assisted-service.company.com/api/assisted-install"
  timeout  = "120s"  # Longer timeout for internal networks
}
```

### Token Rotation

Implement regular token rotation:

1. Generate new offline token monthly
2. Update secret stores/CI systems
3. Test with new token before revoking old token
4. Monitor for authentication failures

## Environment-Specific Configuration

### Development Environment

```hcl
# dev.tfvars
provider "oai" {
  timeout = "300s"  # Longer timeout for slower development environments
}
```

### Production Environment

```hcl
# prod.tfvars
provider "oai" {
  timeout = "30s"   # Strict timeout for production reliability
}
```

### Multi-Environment Setup

```hcl
locals {
  config = {
    dev = {
      timeout = "300s"
    }
    prod = {
      timeout = "30s"
    }
  }
}

provider "oai" {
  timeout = local.config[var.environment].timeout
}
```

## Troubleshooting Authentication

### Common Issues

#### Token Expired
**Symptoms**: HTTP 401 errors, authentication failed messages
**Solution**: Generate new offline token from Red Hat console

#### Network Connectivity
**Symptoms**: Timeouts, connection refused errors
**Solution**: 
- Verify network access to api.openshift.com
- Check corporate firewall/proxy settings
- Increase timeout values

#### Invalid Token Format
**Symptoms**: HTTP 400 errors, malformed token messages
**Solution**: 
- Verify token was copied completely
- Check for extra whitespace or characters
- Regenerate token if necessary

### Debugging Authentication

Enable detailed logging:

```bash
export TF_LOG=DEBUG
terraform apply
```

Verify token validity:

```bash
curl -H "Authorization: Bearer $OFFLINE_TOKEN" \
  https://api.openshift.com/api/assisted-install/v2/openshift-versions
```

### Token Management Script

Create a helper script for token management:

```bash
#!/bin/bash
# check-token.sh

TOKEN_FILE="$HOME/.openshift-token"

if [[ -f "$TOKEN_FILE" ]]; then
  OFFLINE_TOKEN=$(cat "$TOKEN_FILE")
  
  # Test token validity
  if curl -s -f -H "Authorization: Bearer $OFFLINE_TOKEN" \
     https://api.openshift.com/api/assisted-install/v2/openshift-versions > /dev/null; then
    echo "✅ Token is valid"
    export OFFLINE_TOKEN
  else
    echo "❌ Token expired or invalid"
    echo "Please update token at: https://console.redhat.com"
    exit 1
  fi
else
  echo "❌ No token file found at $TOKEN_FILE"
  echo "Please obtain token from: https://console.redhat.com"
  exit 1
fi
```

Usage:
```bash
source ./check-token.sh
terraform apply
```

This authentication guide ensures secure and reliable access to the OpenShift Assisted Service API for your Terraform deployments.