# AWX Configuration (AWX as Code)

`configure_awx.yml` provisions AWX objects idempotently using the `awx.awx` Ansible collection.
Run it once after AWX is up to create the entire configuration from scratch.

## What it creates

- **1 Organization** — "Home Platform"
- **1 Custom Credential Type** — "Home Platform Vault" (injects all `vault_*` variables)
- **2 Standard Credentials** — SSH key for Machine + Source Control
- **3 Execution Environments** — base-ee, docker-ee, k8s-ee
- **1 Inventory** — "Home Platform" with `home-server` host and 4 groups
- **1 Project** — pointing to this repo via Git SCM
- **13 Job Templates** — full matrix of Infra / Deploy / Ops jobs
- **3 Schedules** — health check (15 min), backup (daily), maintenance (weekly)
- **2 Workflow Templates** — Bootstrap Host, Full Stack Deploy

## Usage

```bash
# Install dependencies
pip install awxkit
ansible-galaxy collection install awx.awx

# Run full configuration
ansible-playbook awx_config/configure_awx.yml \
  -e "awx_host=http://localhost:8052" \
  -e "awx_password=<admin-password>"

# Selective tags
ansible-playbook awx_config/configure_awx.yml --tags credentials
ansible-playbook awx_config/configure_awx.yml --tags templates
ansible-playbook awx_config/configure_awx.yml --tags schedules
ansible-playbook awx_config/configure_awx.yml --tags workflows
```

All tasks are idempotent — re-running applies drift corrections without duplication.

## Secrets

After running `configure_awx.yml`, navigate to:
**AWX UI → Credentials → Home Platform Vault** and fill in all secret fields.

The credential type uses AWX's `extra_vars` injector, so all `vault_*` variables are
automatically available to any job template that attaches the credential.
