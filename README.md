# Template Repository

## Files

### `.golangci.yml`

The configuration of the main linter in use. Visit the [golangci](https://golangci-lint.run/) website for more details.

### `.goreleaser.yml`

The configuration of the release tool in use. Visit the [goreleaser](https://goreleaser.com/) website for more details.

### `renovate.json`

The configuration of the dependency monitoring bot. Visit the [renovate](https://docs.renovatebot.com/) website for more details.

## Directories

### `/.github`

Directory for GitHub specific configuration, but GitHub Actions and Issue Templates.

### `/ci`

This directory contains the sources for the CI helper tool which is based on Dagger. It allows easier reproduction of CI results locally.

---

The following general directory structure recommendations are taken from the [project layout](https://github.com/golang-standards/project-layout)

### `/cmd`

Main applications for this project.

The directory name for each application should match the name of the executable you want to have (e.g., `/cmd/myapp`).

Don't put a lot of code in the application directory. If you think the code can be imported and used in other projects, then it should live in the `/pkg` directory. If the code is not reusable or if you don't want others to reuse it, put that code in the `/internal` directory. You'll be surprised what others will do, so be explicit about your intentions!

It's common to have a small `main` function that imports and invokes the code from the `/internal` and `/pkg` directories and nothing else.

### `/pkg`

Library code that's ok to use by external applications (e.g., `/pkg/mypubliclib`). Other projects will import these libraries expecting them to work, so think twice before you put something here :-) Note that the `internal` directory is a better way to ensure your private packages are not importable because it's enforced by Go. The `/pkg` directory is still a good way to explicitly communicate that the code in that directory is safe for use by others. The [`I'll take pkg over internal`](https://travisjeffery.com/b/2019/11/i-ll-take-pkg-over-internal/) blog post by Travis Jeffery provides a good overview of the `pkg` and `internal` directories and when it might make sense to use them.

### `/internal`

Private application and library code. This is the code you don't want others importing in their applications or libraries. Note that this layout pattern is enforced by the Go compiler itself. See the Go 1.4 [`release notes`](https://golang.org/doc/go1.4#internalpackages) for more details. Note that you are not limited to the top level `internal` directory. You can have more than one `internal` directory at any level of your project tree.

### `/tools`

Supporting tools for this project. Note that these tools can import code from the `/pkg` and `/internal` directories.
