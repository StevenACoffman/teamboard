### Teamboard - Phabricator Dashboard, but for Github PRs

This is a small local webserver that will make queries to tell you what open reviews
linger for you, your teammates.

This will query the open PRs (authored, mentioning, or review requested) you, your team, or your teamMates.

The trick is, if you have multiple orgs, or multiple teams (likely), you need to pick which.

Right now, it just picks `Khan` and `districts` which is suboptimal.

You should change that in [server.go#L135](https://github.com/StevenACoffman/teamboard/blob/main/pkg/server/server.go#L135).

Any static assets placed in the [pkg/assets](https://github.com/StevenACoffman/teamboard/tree/main/pkg/assets) folder will be served as `static/assets`
So the [team-pr-template.html](https://github.com/StevenACoffman/teamboard/blob/main/pkg/assets/team-pr-template.html) will be available as:
[localhost:3000/static/assets/team-pr-template.html](http://localhost:3000/static/assets/team-pr-template.html).

If you export the environment variable `PORT`, instead of the default `3000`, whatevfer value you set will be used.

### Mage

Instead of `make` and `Makefile`, I used [mage](https://magefile.org/) and made a [magefile](https://github.com/StevenACoffman/teamboard/blob/main/magefile.go).

If you do `brew install mage` then you can run here:
+ `mage -v run` - will run the webserver by doing `go run main.go`
+ `mage generate` - will re-generate the genqlient code by doing `go generate ./...`
+ `mage install` - will build and install the teamboard application

Or just run the go commands by hand.