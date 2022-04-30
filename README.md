# gh projects

[GitHub CLI] extension for [projects (beta)].
As of [v2.9.0](https://github.com/cli/cli/releases/v2.9.0), `gh` already has some
support for projects but may not support projects (beta) until the APIs have stabilized.

## Install

Make sure you have version 2.0 or [newer] of the [GitHub CLI] installed.

```bash
gh extension install heaths/gh-projects
```

### Upgrade

The `gh extension list` command shows if updates are available for extensions. To upgrade, you can use the `gh extension upgrade` command:

```bash
gh extension upgrade heaths/gh-projects

# Or upgrade all extensions:
gh extension upgrade --all
```

## Commands

### edit

Edit a project:

```bash
gh projects edit 1 -d "A short description" --public
gh projects edit 1 --add-issue 4 --add-issue 8
```

### list

List projects:

```bash
gh projects list
gh projects list --search "launch"
```

### view

View a project:

```bash
gh projects view 1
```

## License

Licensed under the [MIT](LICENSE.txt) license.

[GitHub CLI]: https://github.com/cli/cli
[newer]: https://github.com/cli/cli/releases/latest
[projects (beta)]: https://docs.github.com/en/issues/trying-out-the-new-projects-experience/about-projects
