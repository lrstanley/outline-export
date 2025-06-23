<!-- template:define:options
{
  "nodescription": true
}
-->
![logo](https://liam.sh/-/gh/svg/lrstanley/outline-export?icon=logos%3Amarkdown&icon.height=80&bg=topography&bgcolor=rgba(2%2C+0%2C+26%2C+1)&layout=left)

<!-- template:begin:header -->
<!-- do not edit anything in this "template" block, its auto-generated -->
<!-- template:end:header -->

<!-- template:begin:toc -->
<!-- do not edit anything in this "template" block, its auto-generated -->
<!-- template:end:toc -->

## :grey_question: Why

**outline-export** is a tool to export your [Outline](https://getoutline.com) collections/documents to
a variety of formats. You can do this manually, but it's helpful to do on an automated basis.

## :computer: Installation

Check out the [releases](https://github.com/lrstanley/outline-export/releases) page for prebuilt
versions.

### :whale: Container Images (ghcr)

```console
$ mkdir -vp backups/
$ docker run -it --rm --env-file .env -v $PWD/backups:/backups ghcr.io/lrstanley/outline-export:latest \
    --url "https://outline.example.com" \
    --export-path "/backups/" \
    --extract \
    --format markdown
```

### :toolbox: Build From Source

Dependencies (to build from source only):

- [Go](https://golang.org/doc/install) (latest)
- [Make](https://www.gnu.org/software/make/) (version doesn't really matter)

Setup:

```console
git clone https://github.com/lrstanley/outline-export.git && cd outline-export
make
./outline-export --help
```

## :gear: Usage

See [USAGE.md](USAGE.md) for cli flags and environment variables for full usage.

Behind the scenes, this invokes the Outline API, and does the following:

1. Fetch current exports (if any) created within the last 1 hour, that match the format we're expecting.
2. If no exports are found, create a new export.
3. Wait until the export is ready, then download the export.
4. If `--extract` is true, we extract the export zip, serialize all file names, and write into the target
   directory.
5. Once completed, we clean up any exports that were created in the last 1 hour, that match the format
   we're expecting (to ensure we're not creating a bunch of exports that are left around).


### :hammer: Generating a Token

To generate a token, go to "Settings" > "API & Apps" > "New API Key".

The following scopes are required, however, you can also leave the box blank just in case any others
are needed in the future:

```
collections.export_all fileOperations.info fileOperations.list fileOperations.redirect fileOperations.delete
```

### :bulb: Examples

Export all documents in a collection to markdown files in a directory:

```bash
$ export TOKEN="1234567890"
$ outline-export \
    --url "https://outline.example.com" \
    --export-path "your-export-path/" \
    --extract \
    --format markdown
```

Export the entire backup zip for archival purposes:

```bash
$ export TOKEN="1234567890"
$ outline-export \
    --url "https://outline.example.com" \
    --export-path "outline-backup-$(date +%Y-%m-%d).zip" \
    --format zip
```

<!-- template:begin:support -->
<!-- do not edit anything in this "template" block, its auto-generated -->
<!-- template:end:support -->

<!-- template:begin:contributing -->
<!-- do not edit anything in this "template" block, its auto-generated -->
<!-- template:end:contributing -->

<!-- template:begin:license -->
<!-- do not edit anything in this "template" block, its auto-generated -->
<!-- template:end:license -->
