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
import pytest

from commons.constants import CREDENTIAL, DISCOVERY_PROCESSOR_ENTITY, ENDPOINT, INGESTION_PROCESSOR_ENTITY, PIPELINE, \
  SCHEDULER, SEED
from commons.custom_classes import PdpEntity


@pytest.mark.parametrize('entity_type ,referenced_info', [
  (DISCOVERY_PROCESSOR_ENTITY, {}),
  (ENDPOINT, {DISCOVERY_PROCESSOR_ENTITY.reference_field: DISCOVERY_PROCESSOR_ENTITY}),
  (PIPELINE, {INGESTION_PROCESSOR_ENTITY.reference_field: INGESTION_PROCESSOR_ENTITY}),
  (SEED, {
    PIPELINE.reference_field: PIPELINE,
    CREDENTIAL.reference_field: CREDENTIAL
  }),
  (SCHEDULER, {SEED.reference_field: SEED}),
  (CREDENTIAL, {})
])
def test_get_references(entity_type: PdpEntity, referenced_info):
  """
  Test the method of the class PdpEntity defined in :func:`commons.custom_classes.PdpEntity.get_references`.
  """
  assert entity_type.get_references() == referenced_info
