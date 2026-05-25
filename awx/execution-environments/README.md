# Execution Environments

AWX runs playbooks inside container images called Execution Environments (EEs).
Three EEs are defined here, each extending the previous:

| EE | Collections | Extra tools | Use for |
|----|-------------|-------------|---------|
| `base-ee` | ansible.posix, community.general, ansible.utils | — | OS provisioning, info gathering |
| `docker-ee` | + community.docker | Docker CLI | Docker Compose deploys, container ops |
| `k8s-ee` | + kubernetes.core | kubectl, k3d, helm | Kubernetes / k3d deployments |

## Build

Requires [ansible-builder](https://ansible-builder.readthedocs.io/en/stable/) >= 3.0:

```bash
pip install ansible-builder
bash ../awx_config/build_execution_environments.sh
```

Or build individually:

```bash
cd docker-ee
ansible-builder build \
  --file execution-environment.yml \
  --tag localhost/home-platform/docker-ee:latest \
  --container-runtime docker
```

## Register in AWX

After building, register each EE in AWX:

```bash
ansible-playbook ../awx_config/configure_awx.yml --tags ee \
  -e "awx_password=<password>"
```

Or manually in the AWX UI: **Administration → Execution Environments → Add**.
