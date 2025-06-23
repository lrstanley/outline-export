## :gear: Usage

#### Application Options
| Environment vars | Flags | Type | Description |
| --- | --- | --- | --- |
| `URL` | `--url` | string | URL of the Outline server [**required**] |
| `TOKEN` | `--token` | string | Token for the Outline server [**required**] |
| `FORMAT` | `--format` | string | Format of the export [**required**] [**choices: markdown, html, json**] |
| `EXCLUDE_ATTACHMENTS` | `--exclude-attachments` | bool | Exclude attachments from the export |
| `EXTRACT` | `--extract` | bool | Extract the export into the target directory |
| `EXPORT_PATH` | `--export-path` | string | Path to export the file to. If extract is enabled, this will be the directory to extract the export to. [**required**] |
| `FILTER` | `--filter` | []string | Filters the export to only include certain files (when using --extract). This is a glob pattern, and it matches the files/folders inside of the export zip, not necessarily collections/document exact names. Can be specified multiple times. |
| - | `-v, --version` | bool | prints version information and exits |
| - | `--version-json` | bool | prints version information in JSON format and exits |
| `DEBUG` | `-D, --debug` | bool | enables debug mode |

#### Logging Options
| Environment vars | Flags | Type | Description |
| --- | --- | --- | --- |
| `LOG_QUIET` | `--log.quiet` | bool | disable logging to stdout (also: see levels) |
| `LOG_LEVEL` | `--log.level` | string | logging level [**default: info**] [**choices: debug, info, warn, error, fatal**] |
| `LOG_JSON` | `--log.json` | bool | output logs in JSON format |
| `LOG_GITHUB` | `--log.github` | bool | output logs in GitHub Actions format |
| `LOG_PRETTY` | `--log.pretty` | bool | output logs in a pretty colored format (cannot be easily parsed) |
| `LOG_PATH` | `--log.path` | string | path to log file (disables stdout logging) |
