# Fluid Project Governance 

## Goals
* Support Fluid in achieving its project long term and short term goals 
* Maintain project agility in key technical areas by defining clear lines of responsibility and ownership and by outlining clear processes for resolving technical and non-technical disagreements
* Foster a clear roadmap and value proposition that organizes development work
* Grow a strong "surface area" ecosystem 
  - Support breadth of use cases and integration points
  - Provide room for innovation in a fast-moving field
* Provide a clear and stable path for deep OSS collaboration with partner organizations (downstream and upstream projects, end users, and contributing companies)
* Ensure project has sustained investment and resources it needs to succeed 

# Governance Model
The following is a high-level description of our proposed governance model: it is not an exhaustive list of rules for operation, and exact nomenclature is not set in stone. 

## Groups and Leadership
This section outlines the mechanisms of collaboration within the project. 

###  Role:

**Contributors**: comments on an issue or pull request, people who add value to the project (whether it’s triaging issues, writing code, or organizing events), or anybody with a merged pull request (perhaps the narrowest definition of a contributor).

**Committers**: Committers are community members who have shown that they are committed to the continued development of the project through ongoing engagement with the community. 

**Current Committers:**  are  here: [https://github.com/fluid-cloudnative/fluid/blob/master/MAINTAINERS_COMMITTERS.md](https://github.com/fluid-cloudnative/fluid/blob/master/MAINTAINERS_COMMITTERS.md)

Committers’ responsibility:
* Are expected to work on public branches of the source repository and submit pull requests from that branch to the master branch.
* Must submit pull requests for all changes.
* Have their work reviewed by maintainers before acceptance into the repository.
* Review the PR work from other people in the community.

How to become a new Committer:
* One must have shown a willingness and ability to participate in the project as a team player. Typically, a potential Committer will need to show that they have an understanding of and alignment with the project, its objectives, and its strategy.
* Committers are expected to be respectful of every community member and to work collaboratively in the spirit of inclusion.
* Have submitted a minimum of 5 qualifying pull requests. What’s a qualifying pull request? One that carries significant technical weight and requires little effort to accept because it’s well documented and tested.

Process for Becoming Committers
1. Nominated by one of the maintainers or TOC members by open a community issue
2. Have more than two +1 vote from maintainer or TOC member
3. Community chair send invitation letters, and add the GitHub user to the “Fluid” team

<font size=5>A Committer is invited to become a maintainer by existing maintainers. A nomination will result in discussion and then a decision by the maintainer team.</font>

**Maintainer**: Maintainers are expected to contribute increasingly complicated PRs/designs and review PRs/designs, under the guidance of the existing maintainers. One who wants to be a maintainer should have been working for this project for 3 months at least. New maintainers are nominated and elected by the project maintainers with 2/3 majority vote pass

**Current Maintainers:**  are  here: [https://github.com/fluid-cloudnative/fluid/blob/master/MAINTAINERS_COMMITTERS.md](https://github.com/fluid-cloudnative/fluid/blob/master/MAINTAINERS_COMMITTERS.md)

Maintainers’ responsibilities:

&emsp;Fulfill all responsibilities of Committers, and also:
* Curate github issues and review pull requests / designs for other maintainers and the community.
* Maintainers are expected to respond to assigned Pull Requests in a reasonable time frame.
* A large Pull Request(More than 20 files+ and 2000 lines+ changes) should have at least 3 Maintainers' +1 vote
* Participate when called upon in the security release process. Note that although this should be a rare occurrence, if a serious vulnerability is found, the process may take up to several full days of work to implement.
* In general continue to be willing to spend at least 20% of your time working on Fluid (1 day per week).
* Answer the questions put forward by the community users in maillist, slack, DingTalk, WechatGroup, etc.

To become a new Maintainer:
* Work in a helpful and collaborative way with the community.
* Have given good feedback on others’ submissions and displayed an overall understanding of the code quality standards for the project.
* Commit to being a part of the community for the long-term.
* Have submitted a minimum of 15 qualifying pull requests.

Process to become a Maintainer:
* Talk to one of the existing TOC members to show your interest in becoming a maintainer. Becoming a maintainer generally means that you are going to be spending substantial time (>20% a week) on Fluid for the foreseeable future.
* the TOC member will send the nomination email for the committer to introduce the his/her contribution. the TOC shall hold a vote on this. These votes can happen on the phone, email, or via a voting service, when appropriate. TOC members can either respond "agree, yes, +1", "disagree, no, -1", or "abstain". A vote passes with two-thirds vote of votes
* Community chair send invitation letters, and add the Github user to maintainer page

**Technical Oversight Committee（TOC）**:  The TOC functions as the core management team that oversees the community. The TOC has additional responsibilities over and above those of Maintainers. These responsibilities ensure the smooth running of the project.

Members of the TOC do not have significant authority over other members of the community, although it is the TOC to decide that whether votes on new Maintainers or Committers is valid within 2 weeks of Maintainers or Committers onboarded and could recall it if it's invalid, and makes all major decisions for the future with respect to Fluid, such as project-level governance policies, management of sub-structures, security processes and so on. It also makes decisions when community consensus cannot be reached. In addition, the TOC has access to the project’s private mailing list and its archives.

TOC Members’ responsibilities:

&emsp;Fulfill all responsibilities of Maintainers, and also:
* Curate github issues and review pull requests / designs for other maintainers and the community.
* In general continue to be willing to spend at least 20% of your time working on Fluid (1 day per week).

To become a new TOC Member:

Membership of the TOC is by invitation from the existing TOC members. A nomination will result in discussion and then a vote by the existing TOC members. TOC membership votes are subject to consensus approval of the current TOC members.

For the first formation of the committee, our maintainers nominated a group of people who had great impact or influence on the project, including some members from Alibaba, Alluxio and Nanjing University who founded Fluid. New TOC members are selected based on the rules described above.

In various situations the TOC shall hold a vote. These votes can happen on the phone, email, or via a voting service, when appropriate. TOC members can either respond "agree, yes, +1", "disagree, no, -1", or "abstain". A vote passes with two-thirds vote of votes cast based on the charter. An abstain vote equals not voting at all.

### Membership 
- Initial membership:
   - Community Chair: Rong Gu (Nanjing University) 
   - TOC Members:
      - Bin Fan (Alluxio)
      - Rong Gu (Nanjing University)
      - Kai Zhang （Alibaba Cloud）
      - Yang Che (Alibaba Cloud)

**Community Chair**: Community chair is primarily responsible for performing community development work and administrative functions.

Community Chair’ responsibility:
* Are expected to work on promoting the project publicly, the work includes but not limited to giving talks, generating papers/blogs, organizing community meetings, recruiting contributors/users, etc.
* Collecting and compiling topics for the community meeting agenda, chairing the meeting, ensuring that quality meeting minutes are published.
* Follow-up community actions tracked and resolved.

How to become the Community Chair:
* One must have shown a willingness and ability to participate in the project as the Community Chair. Typically, the Community Chair will need to show that they have an understanding of and alignment with the project, its objectives, and its strategy, and usually have already been a maintainer.
* The Community Chair is expected to be respectful of every community member and to work collaboratively in the spirit of inclusion.
* Nominated by at least a maintainer. One maintainer can only nominate one community chair candidate once.

Process for Becoming Community Chair
1. Nominated by one of the maintainers.
2. Agreed by the majority of maintainers(1/2 majority maintainers vote pass).
3. The nominator sends announcement letters to the maintainers, and updates the GitHub page.
