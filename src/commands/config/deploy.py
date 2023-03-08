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


def run():
  pass
  # For each product target (Ingestion, Core and Discovery)
  #   Creates a spinner it will inform about the progress for each product and each entity type
  #   For each entity file (in the deployment order)
  #     Read the .json file and add the entity to the product entities context ctx[entities][product]
  #     For each entity ctx[entities][product]
  #       Read each entity and associate the name with the id { name: id }
  #       Replace all the name references with the id of the referenced entity
  #       Calls to create_or_updated_entity within a handle_and_continue handler
  #
