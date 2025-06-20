# 🔐 GitHub PRD Environment Setup Guide

## What is the PRD Environment?

The **PRD (Production) environment** you created is a security feature that allows you to:
- ✅ Store sensitive secrets (Azure credentials) separately
- ✅ Add protection rules for production deployments
- ✅ Track deployment history
- ✅ Require manual approvals before deployments (optional)

## Step-by-Step Setup

### 1. Add Azure Secrets to PRD Environment

**Path**: Repository Settings → Environments → PRD → Environment secrets

Add these two secrets:

| Secret Name | Value | Description |
|-------------|-------|-------------|
| `AZURE_CLIENT_ID` | Your Azure Service Principal ID | Used to authenticate with Azure |
| `AZURE_CLIENT_SECRET` | Your Azure Service Principal Secret | Password for Azure authentication |

### 2. Get Azure Service Principal Credentials

If you don't have Azure credentials yet:

```bash
# Create a service principal for your container registry
az ad sp create-for-rbac \
  --name "coolify-go-github-actions" \
  --role "AcrPush" \
  --scopes "/subscriptions/{subscription-id}/resourceGroups/{resource-group}/providers/Microsoft.ContainerRegistry/registries/shrtso"

# Output will show:
{
  "appId": "12345678-1234-1234-1234-123456789abc",     # This is AZURE_CLIENT_ID
  "displayName": "coolify-go-github-actions",
  "password": "super-secret-password-here",             # This is AZURE_CLIENT_SECRET
  "tenant": "87654321-4321-4321-4321-cba987654321"
}
```

### 3. Optional: Add Protection Rules

**Path**: Repository Settings → Environments → PRD → Protection rules

You can add:
- ✅ **Required reviewers** - Require someone to approve deployments
- ✅ **Wait timer** - Add delay before deployment
- ✅ **Deployment branches** - Only allow deployments from main/v4.x

### 4. Test the Setup

1. **Push a change** to `go-src/` directory
2. **Check Actions tab** - Should see workflow running
3. **View deployment** in Environments → PRD → View deployment

## Environment vs Repository Secrets

| Type | Location | Security | Use Case |
|------|----------|----------|----------|
| **Environment Secrets** | PRD Environment | 🔒 High | Production deployments |
| **Repository Secrets** | Repository Settings | 🔓 Medium | Development/testing |

Your **PRD environment** is more secure because:
- Secrets are only available during production deployments
- You can add approval requirements
- Deployment history is tracked
- You can restrict which branches can deploy

## Current Workflow Behavior

With `environment: PRD` in the workflow:

1. **Pull Requests** → Build only (no secrets access)
2. **Main branch pushes** → Deploy to PRD environment with Azure secrets
3. **Version tags** → Deploy to PRD environment with version updates

## Troubleshooting

### If deployment fails with "Environment not found":
- Make sure the PRD environment exists in repository settings
- Check that secrets are added to the **PRD environment**, not repository secrets

### If Azure login fails:
- Verify `AZURE_CLIENT_ID` and `AZURE_CLIENT_SECRET` are correct
- Test credentials locally: `az login --service-principal -u $AZURE_CLIENT_ID -p $AZURE_CLIENT_SECRET --tenant $TENANT_ID`

### If you want manual approval:
- Go to PRD environment settings
- Add yourself as "Required reviewer"
- Now deployments will wait for your approval

Your PRD environment setup provides enterprise-grade security for production deployments! 🛡️
