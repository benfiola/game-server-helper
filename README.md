# game-server-helper

This project aims to reduce the amount of boilerplate needed to create the docker image entrypoints for game servers that I write.

Boilerplate, in this case, includes:

- Wiring up a basic CLI for the entrypoint
  - Provide hook for health checks (if needed)
  - Provide hook for entrypoint
  - Provide print version command
- Performing privilege de-escalation as a bootstrapping step
  - Determining the desired non-root UID/GID
  - Updating a local user to use this UID/GID
  - Taking ownership of necessary directories with this local user
  - Relaunching the entrypoint as this user
- Exposing common operations
  - Downloading urls
  - Extracting archives
  - Running commands
  - Creating and taking ownership of directories
  - Creating symlinks
  - Handling signals

## Installation

```shell
go get github.com/benfiola/game-server-helper
```

## Minimal Example

```golang
package main

import "github.com/benfiola/game-server-helper/pkg/helper"

func Entrypoint(ctx context.Context, api helper.Api) error {
    api.Logger.Info("this is the entrypoint")
    return nil
}

func main() {
    (&helper.Entrypoint{
        Dirs: map[string]string{
            "server": "/server",
            "data": "/data",
        },
        Main: Entrypoint,
        Version: "0.0.0",
    }).Run()
}
```
