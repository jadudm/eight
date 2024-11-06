local B = import 'base.jsonnet';

local fetch_parameters = [
  [
    'workers',
    { cf: 10, container: 10 },
  ],
  [
    'polite_sleep',
    { cf: 2, container: 2 },
  ],
  [
    'polite_cache_default_expiration',
    { cf: B.hours(10), container: B.minutes(10) },
  ],
  [
    'polite_cache_cleanup_interval',
    { cf: B.hours(3), container: B.minutes(5) },
  ],
];

{
  fetch_parameters: [["fetch"] + x for x in fetch_parameters],
  fetch_cf: B.parameters('cf', 'fetch', self.fetch_parameters),
  fetch_container: B.parameters('container', 'fetch', self.fetch_parameters)
}
