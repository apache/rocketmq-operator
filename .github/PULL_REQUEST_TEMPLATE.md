## What is the purpose of the change

XXXXX

## Brief changelog

XX

## Verifying this change

XXXX

**Please go through this checklist to help us incorporate your contribution quickly and easily.**

Notice: `It would be helpful if you could finish the following checklist (the last one is not necessary) before request the community to review your PR`.

- [ ] Make sure there is a [Github issue](https://github.com/apache/rocketmq-operator/issues) filed for the change (usually before you start working on it). Trivial changes like typos do not require a Github issue. Your pull request should address just this issue, without pulling in other changes - one PR resolves one issue. 
- [ ] Format the pull request title like `[ISSUE #123] Fix UnknownException when host config not exist`. Each commit in the pull request should have a meaningful subject line and body.
- [ ] Write a pull request description that is detailed enough to understand what the pull request does, how, and why.
- [ ] Check RBAC rights for Kubernetes roles.
- [ ] Write necessary unit-test to verify your logic correction, more mock a little better when cross module dependency exist. 
- [ ] Run `make docker-build` to build docker image for operator, try your changes from Pod inside your Kubernetes cluster, **not just locally**. Also provide screenshots to show that the RocketMQ cluster is healthy after the changes. 
- [ ] Before committing your changes, remember to run `make manifests` to make sure the CRD files are updated. 
- [ ] Update documentation if necessary.
- [ ] If this contribution is large, please file an [Apache Individual Contributor License Agreement](http://www.apache.org/licenses/#clas).
