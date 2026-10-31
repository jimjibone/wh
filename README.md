# wh - woodhouse client module

This repository contains the Woodhouse client Go module.

The module code is stored in a major version directory matching the
woodhouse-api version it implements. Only the current version is kept: version
2 replaced version 1 outright.

## Version 2

This is the current and only version of the module and all files can be found
under `v2/`. Examples are under `examples/v2-bridge` and `examples/v2-reactor`.

### Generated code

The typed services are generated from the woodhouse-api service schema
(`schema/services.yaml`) by `v2/internal/whgen`. Run `task generate` (or
`go generate ./...`) after moving to a woodhouse-api version with a changed
schema. It writes:

- `v2/types/zz_types.go`: params, results, event fields and list/struct items.
- `v2/bridges/services/zz_services.go`: one typed bridge service per service
  type (`services.NewLight`, ...), with typed attributes, `On<Command>`
  handlers and `Emit<Event>` methods.
- `v2/reactors/zz_views.go`: one typed reactor view per service type
  (`rc.Light(device, endpoint)`, ...), with getters, `Set<Attr>` writes,
  command methods and `On<Event>` handlers.

The woodhouse client code takes care of connecting to the woodhouse-core server.
After the woodhouse-core user has approved the pairing of the client it will be
authenticated and be able to interact with the API (via woodhouse-api). This
enables the client to act as a bridge (providing device updates and control) and
to act as a reactor (receiving updates and enabling control of devices, i.e.
automation).

## History

If you go looking through the git commit history you will notice that some
commits look slightly odd. This is because the files were extracted while
retaining commit history from the original monorepo which is now known as
woodhouse-core. This is to allow for future reasoning about why things are the
way they are. Git commit messages can be very useful if written well.
