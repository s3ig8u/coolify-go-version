# 🔐 Fix "Username and password required" Error

## The Problem
The workflow is failing because the Azure Container Registry credentials are missing from your PRD environment.

## Step-by-Step Fix

### 1. Add Secrets to PRD Environment

**Navigate to**: Your GitHub Repository → Settings → Environments → PRD → Environment secrets

**Add these two secrets:**

| Secret Name | Value | Where to Get It |
|-------------|-------|------------------|
| `AZURE_CLIENT_ID` | Service Principal App ID | See instructions below |
| `AZURE_CLIENT_SECRET` | Service Principal Password | See instructions below |

### 2. Get Azure Service Principal Credentials

#### Option A: Create New Service Principal (Recommended)
```bash
# Login to Azure
az login

# Create service principal with ACR push permissions
az ad sp create-for-rbac \
  --name "coolify-go-github-actions" \
  --role "AcrPush" \
  --scopes "/subscriptions/YOUR_SUBSCRIPTION_ID/resourceGroups/YOUR_RESOURCE_GROUP/providers/Microsoft.ContainerRegistry/registries/shrtso"

# Example output:
{
  "appId": "12345678-1234-1234-1234-123456789abc",        # ← This is AZURE_CLIENT_ID
  "displayName": "coolify-go-github-actions",
  "password": "AbC123XyZ789-SuperSecretPassword",           # ← This is AZURE_CLIENT_SECRET
  "tenant": "87654321-4321-4321-4321-cba987654321"
}
```

#### Option B: Use Existing Service Principal
If you already have a service principal:
```bash
# List existing service principals
az ad sp list --display-name "your-app-name"

# Reset password if needed
az ad sp credential reset --id "your-app-id"
```

#### Option C: Use Registry Admin User (Quick but less secure)
```bash
# Enable admin user on your registry
az acr update --name shrtso --admin-enabled true

# Get admin credentials
az acr credential show --name shrtso

# Use the output:
# AZURE_CLIENT_ID = username (usually same as registry name)
# AZURE_CLIENT_SECRET = password or password2
```

### 3. Verify Secrets are Added

**Check**: Repository Settings → Environments → PRD → Environment secrets

You should see:
- ✅ `AZURE_CLIENT_ID` 
- ✅ `AZURE_CLIENT_SECRET`

### 4. Test Locally (Optional)
```bash
# Test the credentials work
docker login shrtso.azurecr.io \
  --username "YOUR_AZURE_CLIENT_ID" \
  --password "YOUR_AZURE_CLIENT_SECRET"

# If successful, you should see: "Login Succeeded"
```

### 5. Re-run the Workflow

**Method 1**: Push a small change
```bash
echo "# Test secrets" >> go-src/README.md
git add go-src/README.md
git commit -m "Test Azure secrets setup"
git push origin main
```

**Method 2**: Re-run failed workflow
- Go to Actions tab → Failed workflow → "Re-run all jobs"

## Common Issues & Solutions

### Issue: "Service principal not found"
**Solution**: Make sure the `appId` from `az ad sp create-for-rbac` is used as `AZURE_CLIENT_ID`

### Issue: "Access denied to registry"
**Solution**: Verify the service principal has `AcrPush` role:
```bash
az role assignment list --assignee "YOUR_AZURE_CLIENT_ID" --scope "/subscriptions/YOUR_SUBSCRIPTION_ID/resourceGroups/YOUR_RESOURCE_GROUP/providers/Microsoft.ContainerRegistry/registries/shrtso"
```

### Issue: "Secrets not found in environment"
**Solution**: Make sure secrets are added to **PRD environment**, not repository secrets

### Issue: "Invalid credentials"
**Solution**: 
1. Double-check the secret values (no extra spaces)
2. Try resetting the service principal password
3. Use admin user as temporary workaround

## Quick Troubleshooting Commands

```bash
# Check if registry exists and is accessible
az acr show --name shrtso

# Test service principal login
az login --service-principal \
  --username "YOUR_AZURE_CLIENT_ID" \
  --password "YOUR_AZURE_CLIENT_SECRET" \
  --tenant "YOUR_TENANT_ID"

# List available registries
az acr list --query "[].{Name:name,ResourceGroup:resourceGroup}" --output table
```

## Expected Success After Fix

Once secrets are properly configured, you should see:
```
✅ Log in to Azure Container Registry
✅ Build and push Docker image
✅ Test installation script
🎉 Workflow completed successfully
```

The key is making sure both `AZURE_CLIENT_ID` and `AZURE_CLIENT_SECRET` are correctly added to your **PRD environment secrets**! 🔐
