# Release Process

## Create a Pull Request

First create a pull request indicating a version bump. That PR shall only contain a single commit stating
that a version bump has happened. Adhere to the [versioning guideline](https://github.com/u-bmc/u-bmc/blob/main/VERSIONING.md)
on how to name the new version. Make sure to bump any constants in the code indicating the current version.

## Tag

Once the Pull Request with all the version changes has been approved and merged it is time to tag the merged commit.

***IMPORTANT***: [There is currently no way to remove an incorrectly tagged version of a Go module](https://github.com/golang/go/issues/34189).
It is critical you make sure the version you push upstream is correct.

1. Create a signed tag for the new version bump commit on the main branch

    ```sh
    git tag -s -a vX.X.X
    ```

    Make sure current `HEAD` of your working directory is the correct commit.

2. Push tags to the upstream remote `github.com/u-bmc/u-bmc.git` and not your fork

    ```sh
    git push upstream <tag name>
    ```

## Release

Once the tag is pushed our CI will make sure a new release is being created, signed and published.
