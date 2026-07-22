# Timetta Link Preview for Mattermost

Server-side Mattermost plugin that recognizes configured Timetta frontend links,
loads the referenced entity through OData with a Bearer token, and adds a compact
message attachment containing the entity code, name, state, and a link back to
Timetta.

## License

Licensed under the [Apache License 2.0](LICENSE).

Example recognized link:

```text
https://app.timetta.com/issue/DEV-4608/main?navigation=my.dev
```

## Build

Requirements:

- Go 1.25 or newer
- PowerShell 7+

Run:

```powershell
./build.ps1
```

The installable bundle is written to:

```text
dist/com.timetta.link-preview-0.3.0.tar.gz
```

The bundle contains Linux (amd64/arm64), macOS (amd64/arm64), and Windows
(amd64) server executables.

## Install

1. Enable plugin uploads in Mattermost (`PluginSettings.EnableUploads=true`).
2. In **System Console → Plugins → Plugin Management**, upload the archive from
   `dist` and enable **Timetta Link Preview**.
3. Open **System Console → Plugins → Timetta Link Preview** and fill in the
   frontend URL, API URL, and Bearer token.

Unsigned custom plugins require `PluginSettings.RequirePluginSignature=false`.

## Supported routes

Frontend routes are mapped to OData collections in `server/routes.go`:

| Frontend route | OData collection |
|---|---|
| `issue` | `Issues` |
| `projects` | `Projects` |
| `deals` | `Deals` |
| `employees` | `Users` |
| `timesheets` | `TimeSheets` |
| `time-off` | `TimeOffRequests` |
| `expenses` | `ExpenseRequests` |
| `organizations` | `Organizations` |

For a Guid identifier the plugin requests:

```text
GET {ApiURL}/OData/{EntitySet}({guid})
```

For a string identifier it requests the collection with an exact OData filter:

```text
GET {ApiURL}/OData/{EntitySet}?$filter={KeyProperty} eq '{key}'&$top=1
```

The card title displays `code — name` when `code` is available, otherwise just
`name`. It displays `state.name` when available. Issue cards additionally display
`type.name`, `priority.name`, and `project.name`. The original message remains
unchanged if Timetta is unavailable, returns an error, or the entity is not found.

## Security note

The configured Bearer token is marked as a secret in Mattermost. The plugin uses
that single token for all previews, so its Timetta permissions should be limited
to data that every intended Mattermost channel member is allowed to see. Redirects
from the configured API are deliberately not followed to avoid forwarding the
token to another host.

## Test

```powershell
go test ./...
```
