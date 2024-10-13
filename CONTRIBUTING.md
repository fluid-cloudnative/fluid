# Contributing to Fluid
Welcome to Fluid! We would greatly appreciate it if you could make any contributions to get Fluid improved. This document aims to give you some basic instructions on how to make your contribution accepted by Fluid.

## Code of Conduct
Before any contribution you'd like to make, please have a look at the [code of conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md), which details how contributors are expected to conduct themselves as part of the fluid-cloudnative Community.

## Filing Issues
Issues might be the most common way to get yourself engaged in Fluid project. No matter if you are a user or a developer, you may have some feedback based on daily use or some great ideas that may better improve Fluid, and here's where "issue" comes in. We'd be very glad to hear your voices, so  feel free to file issues to Fluid.

For instance, there are multiple circumstances where you may want to open an issue, including but not limited to:
- Find a bug and want to report it
- Want help
- Find doc incomplete or hard to understand
- Find test cases that can be improved
- Want a new feature
- Propose a new feature design
- Have some performance issue
- Have some questions about the project
- ...

When filing an issue, please make sure all your sensitive data has been excluded. Sensitive data could be password, secret key, network locations, private business data and anything that may do harm to your privacy. 

## Code Contributions
The Fluid project accepts code contributions via Github pull requests(PR), and this is the only way accepted to apply your changes to the Fluid project.

You should submit a PR If any modifications are going to be applied to the Fluid project, including but not limited to:
- Fix typos
- Fix bugs
- Fix or polish documents
- Prune redundant codes
- Add comments to codes for readability
- Add missing test cases
- Add new features or enhance some feature
- Refactor codes
- ...

The following sections will give you step by step instructions on how to get started making code contributions to Fluid.

### Setting up Development Workspace
We assume you've got a Github ID. If then, all you need to do can be summarized to the following steps:

1. **Fork Fluid repository** 

    Click the "Fork" button that can be found at the up-right corner of the [main page of the Fluid repository](https://github.com/fluid-cloudnative/fluid). After then you'll get a forked repository which you Github user have full access to.

2. **Clone your own repository** 
    
    Use `git clone https://github.com/<your-username>/fluid.git` to clone the forked repository to your local machine.

3. **Set remote upstream**
    ```shell
    cd fluid
    git remote add upstream https://github.com/fluid-cloudnative/fluid.git
    git remote set-url --push upstream no-pushing
    ```

4. **Update local working directory**
    ```shell
    git fetch upstream
    git checkout master
    git rebase upstream/master
    ```
5. **Create new branch**
    ```shell
    git checkout -b <new-branch>
    ```
    Develop and make any changes on the `<new-branch>`. For more information about developing Fluid, see [developer guide](docs/en/dev/how_to_develop.md)

### Developer Certificate Of Origin

The Developer Certificate of Origin (DCO) is a lightweight way for contributors to certify that they wrote or otherwise have the right to submit the code they are contributing to the project.

Contributors to the Fluid project sign-off that they adhere to these requirements by adding a Signed-off-by line to commit messages.

```shell
This is my commit message

Signed-off-by: Random J Developer <random@developer.example.org>
```

Git even has a -s command line option to append this automatically to your commit message:

```shell
git commit -s -m 'This is my commit message'
```

If you have already made a commit and forgot to include the sign-off, you can amend your last commit to add the sign-off with the following command, which can then be force pushed.

```shell
git commit --amend -s
```


### Submitting Pull Requests
Once you've done your work on developing Fluid, you are now ready to submit a PR to the Fluid project.

First of all, push your commits to your forked repository on Github

```shell
git push origin <new-branch>
```

Go to the "Pull requests" tab page under your repository and Click the "New Pull Request" button to submit a PR. Select `<new-branch>`, check your commits, fill in PR title and description and finally click the "Create pull request" button to finish submitting.

To help reviewers better get your purpose, PR title should be descriptive enough but not too long. It's also recommended that you follow the [PR template](.github/PULL_REQUEST_TEMPLATE.md) as your PR description.

### Tracking Your PR
Once you've submitted the PR to the Fluid project, your PR will be reviewed. Please keep tracking the status of your PR, make responses to the reviewers' comments and update your changes if needed to make your PR get accepted.

After at least one approval from reviewers, your PR will soon be merged into Fluid code base. Cheer! You've now contributed to the Fluid project. Thank you for your contribution and you are always welcome!


## Any Help Is Contribution

Though contributions via Github PR is an explicit way to help, we still call for any other help:

- reply to other's issue if you could
- help solve other user's problems
- help review PRs
- take part in discussion about Fluid
- advocate Fluid beyond Github
- write blogs about Fluid
- ...

## Join Our Community as a Member
You are warmly welcome if you'd like to join our community as a member. Together we can make this community even better!

**Some requirements are needed to join our community:**

- Have read [Contributing to Fluid](CONTRIBUTING.md) carefully
- Promise to observe our [code of conduct](code-of-conduct.md)
- Have submitted multiple PRs to Fluid
- Be active in the community, including but not limited to:
    - Open new issues
    - Comment on issues
    - Help review PRs
    - Submit PRs

**How to join:**

You can do it in either of the following two ways:

- Submit a PR to let us know
- Contact directly via the [community channels](https://github.com/fluid-cloudnative/fluid#community)
