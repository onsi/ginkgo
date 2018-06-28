# Contributing to Ginkgo

Your contributions to Ginkgo are essential for its long-term maintenance and improvement.

- Please **open an issue first** - describe what problem you are trying to solve and give the community a forum for input and feedback ahead of investing time in writing code!
- Ensure adequate test coverage:
    - When adding to the Ginkgo library, add unit and/or integration tests (under the `integration` folder).
    - When adding to the Ginkgo CLI, note that there are very few unit tests.  Please add an integration test.
- Update the documentation. Ginko uses `godoc` comments and documentation on the `gh-pages` branch.
  If relevant, please submit a docs PR to that branch alongside your code PR.

Thanks for supporting Ginkgo!

## Setup

Fork the repo, then:

```
git clone git@github.com:<NAME>/ginkgo.git $GO_PATH/src/github.com/<NAME>/ginkgo`
go get github.com/onsi/gomega/...

./before_pr.sh # replace imports to test internal packages
ginkgo -r -p   # ensure tests are green
go vet ./...   # ensure linter is happy
```

## Making the PR
 - make your changes
 - run tests and linter again (see above)
 - undo changes from `./before_pr.sh` via `git checkout .`
 - open PR 🎉
