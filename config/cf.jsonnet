local B = import 'base.jsonnet';
local F = import 'fetch.jsonnet';
{
  // :: means "not visible in the output"
  EIGHT_SERVICES: {
    fetch: F.fetch_cf
    }
}
