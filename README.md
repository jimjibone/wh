# wh - woodhouse client module

This repository contains the Woodhouse client Go module.

The module code is stored in major version directories, with there currently
only being version 1. In the future we may introduce new versions, so this
separation can enable backwards compatibility if we choose to support it.

## Version 1

This is the current and only version of the module and all files can be found
under `v1/`.

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
