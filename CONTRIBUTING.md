# Contributing

Thanks for taking the time to contribute to kwil-db! 

Please follow the guidelines below when contributing. If you have any questions, please feel free to reach out to us on [Discord](https://discord.com/invite/HzRPZ59Kay) or use the [discussions](https://github.com/kwilteam/kwil-db/discussions) feature on GitHub.

## Table of Contents

-   [Discussions](#discussions)
-   [Issues](#issues)
-   [Pull Requests](#pull-requests)
    -   [Commit Messages](#commit-messages)
    -   [Coding Style](#coding-style)
    -   [Pull Request Process](#pull-request-process)
-   [License](#license)

## Discussions

Discussions are for general discussions related to kwil-db. This can be to discuss kwil-db architecture, important tradeoffs / considerations, or any topics that cannot be directly closed by a pull request.

Examples of good discussions:

-   Discussing potental ways to price queries in kwil-db
-   Discussing the tradeoffs of using postgres vs sqlite.
-   Discussing potential procedures for handling consensus failures & changes (and a future issue if the discussion determines we need a feature, bug fix, or documentation change).

Discussions can lead to an issue if they determine that a feature, bug fix, or documentation change is needed.

## Issues

Issues are for reporting bugs, requesting features, requesting repository documentation, or discussing other changes to the kwil-db repository.

For general discussions, or discussions where it is unclear how the discussion would be closed by a pull request, please use the [discussion](https://github.com/kwilteam/kwil-db/discussions) section.

For opening issues, please follow the following guidelines:

-   **Use templates** for creating issues about bugs, feature requests, or documentation requests.
-   **Search** the issue tracker before opening an issue to avoid duplicates.
-   **Be clear & detailed** about what you are reporting.

We strongly recommended submitting an issue before submitting a pull request, especially for pull requests that will require significant effort on your part. This is to ensure that the issue is not already being worked on and that your pull request will be accepted. Some features or fixes are intentionally not included in kwil-db - use issues to check with maintainers and save time!

## Pull Requests

### Commit Messages

kwil-db uses recommended [Go commit messages](https://go.dev/doc/contribute#commit_messages), with breaking changes and deprecations to be noted in the commit message footer. Commits should follow the following format:

```
[Package/File/Directory Name]: [Concise and concrete description of the PR/Issue]

[Longer description of the PR/Issue, if necessary]

[Optional: Issues Taged + Breaking changes/deprecations]
```

For example:

```
cmd/kwil-cli: Add new command to do something

This PR adds a new command to do something. It also adds a new flag to the existing command to do something else.

Resolves #123

BREAKING CHANGE: This PR changes the behavior of the foo command. It now does something else.
```

Breaking changes are any changes that effect the external API of packages that are consumed outside of kwil-db (i.e. proto, core, extensions, and cmd). Changes to internal packages (i.e. deployments, internal, parse, scripts, and test) are not considered breaking changes and do not need to be tagged in the commit footer.

### Coding Style

Please ensure that your contributions adhere to the following coding guidelines:

-   Code should adhere to the official Go [formatting](https://go.dev/doc/effective_go#formatting) guidelines (i.e. uses [`gofmt`](https://pkg.go.dev/cmd/gofmt)).
-   Code must be documented adhering to the Go commentary [guidelines](https://go.dev/doc/effective_go#commentary) and Go Doc [comments](https://go.dev/doc/comment).
-   Code should be tested as much as possible. Tests should be placed in the same package as the code they are testing, in a file named `*_test.go` (e.g. `foo.go` should have a corresponding `foo_test.go`).

### Pull Request Process

1. Fork the repository and create a new branch from `main`.

2. Ensure that your PR is up-to-date with the `main` branch.

```bash
git remote add upstream https://github.com/kwilteam/kwil-db.git
git fetch upstream
git merge upstream/main
```

If you have conflicts, you will need to resolve them before continuing.

Please ensure that all the commits in your git history match the commit message [guidelines](#commit-messages) above. You can use `git rebase -i` to edit your commit history.

3. Ensure that your PR is ready to be merged and all tests pass:

```bash
go mod tidy
go test ./...
```

4. Push your branch to github.

```bash
git add .
git commit -m "Your commit message"
git push origin <branch-name>
```

5. Open a pull request to the `main` branch of the kwil-db repository. Please follow the PR template.

6. Wait for a maintainer to review your PR. If there are any issues, you will be notified and you can make the necessary changes.

## License

By contributing to kwil-db, you agree that your contributions will be licensed under its [Apache 2.0 License](https://www.apache.org/licenses/LICENSE-2.0).
