# Contributing to the Clear Containers tests repository

The Clear Containers tests repository is an open source project licensed under the
[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0).

## Coding Style (Go)

The usual Go style, enforced by `gofmt`, should be used. Additionally, the [Go
Code Review](https://github.com/golang/go/wiki/CodeReviewComments) document
contains a few common errors to be mindful of.


## Certificate of Origin

In order to get a clear contribution chain of trust we use the [signed-off-by
language](https://01.org/community/signed-process)
used by the Linux kernel project.

## Patch format

Beside the Signed-off-by footer, we expect each patch to comply with the
following format:

```
       Subsystem: Change summary (no longer than 75 characters)

       More detailed explanation of your changes: Why and how.
       Wrap it to 72 characters.
       See:
           http://chris.beams.io/posts/git-commit/
       for some more good advice, and the Linux Kernel document:
           https://git.kernel.org/cgit/linux/kernel/git/torvalds/linux.git/tree/Documentation/SubmittingPatches

       Fixes: #nnn

       Signed-off-by: <contributor@foo.com>
```

For example:

```
    pod: Remove token from Cmd structure

    The token and pid data will be hold by the new Process structure and
    they are related to a container.

    Fixes: #123

    Signed-off-by: Sebastien Boeuf <sebastien.boeuf@intel.com>
```

Correct formatting of the PR patches is verified using the [checkcommits](https://github.com/clearcontainers/tests/tree/master/cmd/checkcommits) tool.

Note, that the body of the message should not just be a continuation of the
subject line, and is not used to extend the subject line beyond its length
limit. They should stand alone as complete sentence and paragraphs.

It is recommended that each of your patches fixes one thing. Smaller patches
are easier to review, and are thus more likely to be accepted and merged, and
problems are more likely to be picked up during review.

## Pull requests

We accept [github pull requests](https://github.com/clearcontainers/tests/pulls).

Github has a basic introduction to
[using pull requests](https://help.github.com/articles/using-pull-requests/).

When submitting your Pull Request (PR), treat the Pull Request message the same
you would a patch message, including pre-fixing the title with a subsystem
name. Github by default seems to copy the message from your first patch, which
many times is appropriate, but please ensure your message is accurate and
complete for the whole Pull Request, as it ends up in the git log as the merge
message.

Your pull request may get some feedback and comments, and require some rework.
The recommended procedure for reworking is to rework your branch to a new clean
state and 'force push' it to your github. GitHub understands this action, and
does sensible things in the online comment history. Do not pile patches on
patches to rework your branch. Any relevant information from the github
comments section should be re-worked into your patch set, as the ultimate place
where your patches are documented is in the git log, and not in the github
comments section.

For more information on github 'force push' workflows see this
[blog post](http://blog.adamspiers.org/2015/03/24/why-and-how-to-correctly-amend-github-pull-requests/).

It is perfectly fine for your Pull Request to contain more than one patch - use
as many patches as you need to implement the Request (see the previously
mentioned 'small patch' thoughts). Each Pull Request should only cover one
topic - if you mix up different items in your patches or pull requests then you
will most likely be asked to rework them.

## Reviews

Before your Pull Requests are merged into the main code base, they will be
reviewed. Anybody can review any Pull Request and leave feedback (in fact, it
is encouraged).

We use an 'acknowledge' system for people to note if they agree, or disagree,
with a Pull Request. We utilise some automated systems that can spot common
acknowledge patterns, which include placing any of these at the beginning of a
comment line:

 - LGTM
 - lgtm
 - +1
 - Approve

### Project maintainers

The Clear Containers tests maintainers will be the ones accepting or rejecting any pull request. They are listed in the OWNERS files, and there can be one OWNERS file per directory.

The OWNERS files split maintainership into 2 categories: reviewers and approvers. All approvers also belong to the reviewers list and there must be at least one approval from one member of each list for a pull request to be merged.

Since approvers are also reviewers, they technically can approve a pull request without getting another reviewer's approval. However, it is their due diligence to rely on reviewers and should use their approval power only in very specific cases.

## Issue tracking

To report a bug that is not already documented, please [open an issue in
github](https://github.com/clearcontainers/tests/issues/new) so we all get
visibility on the problem and work toward resolution.

To help the developers resolve your issue, if you are running with Clear Containers,
please include the output from the command below in the issue:

```bash
$ cc-runtime cc-env
```

## Closing issues

The preferred way to close issues is by adding the `Fixes` keyword to your commit message.
Our tooling requires a `Fixes` line in at least one commit message per PR:

```
    pod: Remove token from Cmd structure

    The token and pid data will be hold by the new Process structure and
    they are related to a container.

    Fixes #123

    Signed-off-by: Sebastien Boeuf <sebastien.boeuf@intel.com>
```

Github will then automatically close that issue when parsing the
[commit message](https://help.github.com/articles/closing-issues-via-commit-messages/).
