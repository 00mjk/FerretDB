# Contributing

Thank you for your interest in making FerretDB better!

## Finding something to work on

We are interested in all contributions, big or small, in code or documentation.
But unless you are fixing a very small issue like a typo,
we kindly ask you first to [create an issue](https://github.com/FerretDB/FerretDB/issues/new/choose),
to leave a comment on an existing issue if you want to work on it,
or to [join our Slack chat](./README.md#community) and leave a message for us there.
This way, you will get help from us and avoid wasted efforts if something can't be worked on right now
or someone is already working on it.

You can find a list of good issues for first-time contributors [there](https://github.com/FerretDB/FerretDB/contribute).

## Contributing code

The supported way of developing FerretDB is to modify and run it on the host
(Linux, macOS, or Windows with [WSL2](https://docs.microsoft.com/en-us/windows/wsl/)),
with PostgreSQL and other dependencies running inside Docker Compose.

You will need Go 1.18 as FerretDB extensively uses ([fuzzing](https://go.dev/doc/tutorial/fuzz))
and [generics](https://go.dev/doc/tutorial/generics)).
If your package manager doesn't provide it yet,
please install it from [go.dev](https://go.dev/dl/).

### Cloning the Repository

After [forking FerretDB on GitHub](https://github.com/FerretDB/FerretDB/fork),
you can clone the repository:

```sh
git clone git@github.com:<YOUR_GITHUB_USERNAME>/FerretDB.git
cd FerretDB
git remote add upstream https://github.com/FerretDB/FerretDB.git
```

### Setting up the development environment

To run development commands, you should first install the [`task`](https://taskfile.dev/) tool.
You can do this by changing the directory to `tools` (`cd tools`) and running `go generate -x`.
That will install `task` into the `bin` directory (`bin/task` on Linux and macOS, `bin\task.exe` on Windows).
You can then add `./bin` to `$PATH` either manually (`export PATH=./bin:$PATH` in `bash`)
or using something like (`direnv` (`.envrc` files)[https://direnv.net],
or replace every invocation of `task` with explicit `bin/task`.
You can also [install `task` globally](https://taskfile.dev/#/installation),
but that might lead to the version skew.

With `task` installed, you may do the following:

1. Install development tools with `task init`.
2. Download required Docker images with `task env-pull`.
3. Start the development environment with `task env-up`.
   This will start PostgreSQL and MongoDB containers, filling them with identical sets of test data.
4. Run all tests in another terminal window with `task test`.
5. Start FerretDB with `task run`.
   This will start it in a development mode where all requests are handled by FerretDB, but also routed to MongoDB.
   The differences in response are then logged and the FerretDB response is sent back to the client.
6. Run `mongosh` with `task mongosh`.
   This allows you to run commands against FerretDB.

You can see all available `task` tasks with `task -l`.

### Code Overview

The directory `cmd` provides commands implementation.
Its subdirectory `ferretdb` is the main FerretDB binary; others are tools for development.

The package `tools` uses ["tools.go" approach](https://github.com/golang/go/issues/25922#issuecomment-402918061) to fix tools versions.
They are installed into `bin/` by `cd tools; go generate -x`.

The `internal` subpackages contain most of the FerretDB code:

* `types` package provides Go types matching BSON types that don't have built-in Go equivalents:
  we use `int32` for BSON's int32, but `types.ObjectID` for BSON's ObjectId.
* `fjson` provides converters from/to FJSON for built-in and `types` types.
  FJSON adds some extensions to JSON for keeping object keys in order, preserving BSON type information, etc.
  FJSON is used by `jsonb1` handler/storage.
* `bson` package provides converters from/to BSON for built-in and `types` types.
* `wire` package provides wire protocol implementation.
* `clientconn` package provides client connection implementation.
  It accepts client connections, reads `wire`/`bson` protocol messages, and passes them to `handlers`.
  Responses are then converted to `wire`/`bson` messages and sent back to the client.
* `handlers` handle protocol commands.
  They use `fjson` package for storing data in PostgreSQL in jsonb columns, but they don't use `bson` package –
  all data is represented as built-in and `types` types.

Those packages are tested by "unit" tests that are placed inside those packages.
Some of them are truly hermetic and test only the package that contains them;
you can run those "short" tests with `go test -short` or `task test-unit-short`,
but that's typically not required.
Other unit tests use real databases;
you can run those with `task test-unit` after starting the environment as described above.

We also have a set of "integration" tests in `integration` Go module that uses the Go MongoDB driver
and tests either a running MongoDB-compatible database (such as FerretDB or MongoDB itself)
or in-process FerretDB.
They allow us to ensure compatibility between FerretDB and MongoDB.
You can run them with `task test-integration-ferretdb` for in-process FerretDB
(meaning that integration tests start and stop FerretDB themselves),
`task test-integration-mongodb` for MongoDB running on port 37017 (as in our development environment),
or `task test-integration` to run both in parallel.

Finally, you may run all tests in parallel with `task test`.
If tests fail and the output is too confusing, try running them sequentially by using the commands above.

In general, we prefer integration tests over unit tests,
tests using real databases over short tests
and real objects over mocks.

(You might disagree with our terminology for "unit" and "integration" tests;
let's not fight over it.)

We have an additional integration testing system in another repository: <https://github.com/FerretDB/dance>.

### Code style and conventions

Above everything else, we value consistency in the source code.
If you see some code that doesn't follow some best practice but is consistent,
please keep it that way;
but please also tell us about it, so we can improve all of it.
If, on the other hand, you see code that is inconsistent without apparent reason (or comment),
please improve it as you work on it.

Our code most of the standard Go conventions,
documented on [CodeReviewComments wiki page](https://github.com/golang/go/wiki/CodeReviewComments).
Some of our idiosyncrasies:

1. We use type switches over BSON types in many places in our code.
   The order of `case`s follows this order: <https://pkg.go.dev/github.com/FerretDB/FerretDB/internal/types#hdr-Mapping>
   It may seem random, but it is only pseudo-random and follows BSON spec: <https://bsonspec.org/spec.html>

### Submitting code changes

Before submitting a pull request, please make sure that:

1. Tests are added for new functionality or fixed bugs.
2. Code is regenerated if needed (`task gen`).
3. Code is formatted (`task fmt`).
4. Test pass (`task test`).
5. Linters pass (`task lint`).

## Contributing documentation

Please format documentation with `task docs-fmt`.
