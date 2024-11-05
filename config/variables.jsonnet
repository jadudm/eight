local hours(n) = n * 60 * 60;
local minutes(n) = n * 60;

local extract = 'extract';
local fetch = 'fetch';

local services = [
  [
    extract, "workers", {cf: 10, docker: 10},
    extract, "extract_pdf", {cf: true, docker: true},
    extract, "extact_html", {cf: true, docker: true},
    extract, "walkabout", {cf: true, docker: true},
    fetch,
    'workers',
    { cf: 10, docker: 10 },
    fetch,
    'polite_sleep',
    { cf: 2, docker: 2 },
    fetch,
    'police_cache_default_expiration',
    { cf: hours(10), docker: minutes(10) },
    fetch,
    'police_cache_cleanup_interval',
    { cf: hours(3), docker: minutes(5) },
  ],
];

local params(service, env) = {
  parameters: { [s[1]]: s[2][env] for s in services if s[0] == service},
};

{
  cf: {
    extract: params(extract, 'cf'),
    fetch: params(fetch, 'cf'),
  },
  docker: {
    extract: params(extract, 'docker'),
    fetch: params(fetch, 'docker') ,
  },
}