
local extract = 'extract';


local extract_params = [
  [
    'workers',
    { cf: 10, container: 10 },
  ],
  [
    'extract_pdf',
    { cf: true, container: true },
  ],
  [
    'extact_html',
    { cf: true, container: true },
  ],
  [
    'walkabout',
    { cf: true, container: true },
  ],
];

local services =
  [[fetch] + x for x in fetch_params] +
  [[extract] + x for x in extract_params];

local parameters(env, service) = 
  { parameters: {[s[1]]: s[2][env], for s in services if s[0] == service}};


{
  parameters:: parameters,
}