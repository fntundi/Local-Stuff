# AWX local settings — mounted into awx-web and awx-task
# Extends the default AWX configuration.
# https://docs.ansible.com/automation-controller/latest/html/administration/configure_tower_in_tower.html

import os

# Allow all hosts (restrict to your network in production)
ALLOWED_HOSTS = ['*']

# Session and CSRF
SESSION_COOKIE_AGE = 1800   # 30 minutes

# Logging
LOGGING['loggers']['awx']['level'] = 'INFO'

# Execution environment registry mirror (optional)
# GLOBAL_JOB_EXTRA_VARS = {}

# Default execution environment — set after bootstrapping EEs
# DEFAULT_EXECUTION_ENVIRONMENT = 1
