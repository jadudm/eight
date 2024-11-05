local vars = import 'variables.jsonnet';

{
  local root = self,
  # :: means "not visible in the output"
  creds:: { credentials: { port: vars.container_port}},
  EIGHT_SERVICES: {
    extract: vars.params("fetch", 0),
    fetch: root.creds,
    pack: root.creds,
    serve: root.creds,
    walk: root.creds,

  },
}