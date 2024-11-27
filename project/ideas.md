# Ideas to implement in new projects

- Use wire to dependency injection https://github.com/google/wire/blob/main/_tutorial/README.md
- Use a gateway to external http calls (in adpaters) (see adapter.clients.go):
  - Generate a openapi.yml for each client.
  - Use openapi go client generator to generate the go clients
  - Inject the gateway url
  - Made the gateway proxy to the mapped service.
    - Caddy
    - Hookdeck
    - Hook0
    - ...
