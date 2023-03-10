#  Copyright (c) 2023 Pureinsights Technology Ltd. All rights reserved.
#
#  Permission to use, copy, modify or distribute this software and its
#  documentation for any purpose is subject to a licensing agreement with
#  Pureinsights Technology Ltd.
#
#  All information contained within this file is the property of
#  Pureinsights Technology Ltd. The distribution or reproduction of this
#  file or any information contained within is strictly forbidden unless
#  prior written permission has been granted by Pureinsights Technology Ltd.
import os

from commons.custom_classes import PdpEntity
from commons.file_system import list_directories

# Endpoints
URL_EXPORT_ALL = '{0}/export'

# URLs
INGESTION_API_URL = 'http://localhost:8080'
STAGING_API_URL = 'http://localhost:8081'
CORE_API_URL = 'http://localhost:8082'
DISCOVERY_API_URL = 'http://localhost:8088/admin'

# Must be all in lower case
# Products
INGESTION = 'ingestion'
DISCOVERY = 'discovery'
CORE = 'core'
STAGING = 'staging'
PRODUCTS = [CORE, INGESTION, DISCOVERY, STAGING]

# Entities
SCHEDULER = PdpEntity(INGESTION, 'scheduler', 'cron_jobs.json')
INGESTION_PROCESSOR = PdpEntity(INGESTION, 'processor', 'processors.json')
PIPELINE = PdpEntity(INGESTION, 'pipeline', 'pipelines.json')
SEED = PdpEntity(INGESTION, 'seed', 'seeds.json')
CREDENTIAL = PdpEntity(CORE, 'credential', 'credentials.json')
ENDPOINT = PdpEntity(DISCOVERY, 'endpoint', 'endpoints.json')
DISCOVERY_PROCESSOR = PdpEntity(DISCOVERY, 'processors', 'processors.json', 'processors')
# Must be in order (based on which has fewer dependencies to another entities)
ENTITIES = CREDENTIAL, INGESTION_PROCESSOR, PIPELINE, SEED, SCHEDULER, DISCOVERY_PROCESSOR, ENDPOINT

# Configurations
DEFAULT_CONFIG = {
  INGESTION: INGESTION_API_URL,
  DISCOVERY: DISCOVERY_API_URL,
  CORE: CORE_API_URL,
  STAGING: STAGING_API_URL,
}

# Common messages
WARNING_FORMAT = '[WARNING]: {message}'
ERROR_FORMAT = '[ERROR]: {message}'
EXCEPTION_FORMAT = 'Failed to execute command due to an {exception}:\n{error}'

# Severities
ERROR_SEVERITY = 'error'
WARNING_SEVERITY = 'warning'

TEMPLATE_NAMES = [directory.lower() for directory in list_directories(
  os.path.join(os.path.dirname(__file__), '..', 'commands', 'config', 'templates', 'projects'))]
