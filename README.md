<!-- template:define:options
{
  "nodescription": true
}
-->
![logo](https://liam.sh/-/gh/svg/lrstanley/outline-export?icon=logos%3Amarkdown&icon.height=80&bg=topography&bgcolor=rgba(2%2C+0%2C+26%2C+1)&layout=left)

<!-- template:begin:header -->
<!-- do not edit anything in this "template" block, its auto-generated -->

<p align="center">
  <a href="https://github.com/lrstanley/outline-export/tags">
    <img title="Latest Semver Tag" src="https://img.shields.io/github/v/tag/lrstanley/outline-export?style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/outline-export/commits/master">
    <img title="Last commit" src="https://img.shields.io/github/last-commit/lrstanley/outline-export?style=flat-square">
  </a>




  <a href="https://github.com/lrstanley/outline-export/actions?query=workflow%3Atest+event%3Apush">
    <img title="GitHub Workflow Status (test @ master)" src="https://img.shields.io/github/actions/workflow/status/lrstanley/outline-export/test.yml?branch=master&label=test&style=flat-square">
  </a>


  <a href="https://codecov.io/gh/lrstanley/outline-export">
    <img title="Code Coverage" src="https://img.shields.io/codecov/c/github/lrstanley/outline-export/master?style=flat-square">
  </a>

  <a href="https://pkg.go.dev/github.com/lrstanley/outline-export">
    <img title="Go Documentation" src="https://pkg.go.dev/badge/github.com/lrstanley/outline-export?style=flat-square">
  </a>
  <a href="https://goreportcard.com/report/github.com/lrstanley/outline-export">
    <img title="Go Report Card" src="https://goreportcard.com/badge/github.com/lrstanley/outline-export?style=flat-square">
  </a>
</p>
<p align="center">
  <a href="https://github.com/lrstanley/outline-export/issues?q=is:open+is:issue+label:bug">
    <img title="Bug reports" src="https://img.shields.io/github/issues/lrstanley/outline-export/bug?label=issues&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/outline-export/issues?q=is:open+is:issue+label:enhancement">
    <img title="Feature requests" src="https://img.shields.io/github/issues/lrstanley/outline-export/enhancement?label=feature%20requests&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/outline-export/pulls">
    <img title="Open Pull Requests" src="https://img.shields.io/github/issues-pr/lrstanley/outline-export?label=prs&style=flat-square">
  </a>
  <a href="https://github.com/lrstanley/outline-export/discussions/new?category=q-a">
    <img title="Ask a Question" src="https://img.shields.io/badge/support-ask_a_question!-blue?style=flat-square">
  </a>
  <a href="https://liam.sh/chat"><img src="https://img.shields.io/badge/discord-bytecord-blue.svg?style=flat-square" title="Discord Chat"></a>
</p>
<!-- template:end:header -->

<!-- template:begin:toc -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :link: Table of Contents

  - [Why](#grey_question-why)
  - [Installation](#computer-installation)
    - [Container Images (ghcr)](#whale-container-images-ghcr)
    - [Build From Source](#toolbox-build-from-source)
  - [Usage](#gear-usage)
    - [Generating a Token](#hammer-generating-a-token)
    - [Examples](#bulb-examples)
  - [Support &amp; Assistance](#raising_hand_man-support--assistance)
  - [Contributing](#handshake-contributing)
  - [License](#balance_scale-license)
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
## :raising_hand_man: Support & Assistance

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for
     guidelines on ensuring everyone has the best experience interacting with
     the community.
* :raising_hand_man: Take a look at the [support](.github/SUPPORT.md) document on
     guidelines for tips on how to ask the right questions.
* :lady_beetle: For all features/bugs/issues/questions/etc, [head over here](https://github.com/lrstanley/outline-export/issues/new/choose).
<!-- template:end:support -->

<!-- template:begin:contributing -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :handshake: Contributing

* :heart: Please review the [Code of Conduct](.github/CODE_OF_CONDUCT.md) for guidelines
     on ensuring everyone has the best experience interacting with the
    community.
* :clipboard: Please review the [contributing](.github/CONTRIBUTING.md) doc for submitting
     issues/a guide on submitting pull requests and helping out.
* :old_key: For anything security related, please review this repositories [security policy](https://github.com/lrstanley/outline-export/security/policy).
<!-- template:end:contributing -->

<!-- template:begin:license -->
<!-- do not edit anything in this "template" block, its auto-generated -->
## :balance_scale: License

```
MIT License

Copyright (c) 2025 Liam Stanley <liam@liam.sh>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

_Also located [here](LICENSE)_
<!-- template:end:license -->
