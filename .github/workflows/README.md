# 🚀 CI/CD Workflows

## Build and Push Coolify Go to Azure Container Registry

### Overview
This workflow automatically builds and pushes the Coolify Go Docker image to Azure Container Registry (`shrtso.azurecr.io`) whenever changes are made to the `go-src/` directory.

### Triggers
- **Push to main/v4.x branches** with changes in `go-src/`
- **Git tags** starting with `v*` (e.g., `v1.3.0`, `v2.0.0`)
- **Pull requests** to main branch (build only, no push)

### What it does
1. ✅ **Builds** Docker image from `go-src/Dockerfile`
2. ✅ **Logs in** to Azure Container Registry automatically
3. ✅ **Pushes** to `shrtso.azurecr.io/coolify-go`
4. ✅ **Multi-platform** builds (AMD64 + ARM64)
5. ✅ **Tests** installation script syntax
6. ✅ **Updates** version numbers on tag releases

### Required Secrets
Set these in your GitHub repository **PRD environment** settings:

**Path**: Repository Settings → Environments → PRD → Environment secrets

```
AZURE_CLIENT_ID     = Your Azure service principal ID
AZURE_CLIENT_SECRET = Your Azure service principal password
```

### PRD Environment Benefits
- ✅ **Protection rules** - Require reviews before production deployments
- ✅ **Deployment branches** - Only deploy from specific branches
- ✅ **Environment secrets** - Separate production credentials
- ✅ **Deployment history** - Track all production deployments

### Image Tags Generated
- `latest` - For main branch pushes
- `v1.3.0` - For version tags
- `main` - For main branch
- `v4.x` - For v4.x branch
- `pr-123` - For pull requests

### Usage After CI/CD
Once the workflow runs, your installation script will automatically pull from the registry:

```bash
# This will now pull from shrtso.azurecr.io/coolify-go:latest
curl -fsSL https://raw.githubusercontent.com/s3ig8u/coolify-go-version/main/go-src/install.sh | sudo bash
```

### Manual Trigger
You can also manually trigger builds by:
1. Creating a new tag: `git tag v1.4.0 && git push origin v1.4.0`
2. Pushing to main branch with go-src changes
3. Using GitHub Actions "Run workflow" button

### Monitoring
- Check the "Actions" tab in GitHub for build status
- Each successful build creates a deployment summary
- Failed builds will show detailed error logs

This eliminates the need for manual `az acr login` and `docker push` commands!
