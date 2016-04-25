# Snap Collector Plugin - Apache Mesos

  1. [Contributing Code](#contributing-code)
  2. [Contributing Examples](#contributing-examples)
  3. [Contributing Elsewhere](#contributing-elsewhere)
  4. [Thank You](#thank-you)

This repository has dedicated developers from Intel working on updates. The most helpful way to contribute is by
reporting your experience through issues. Issues may not be updated while we review internally, but they're still
incredibly appreciated.

## Contributing Code
**_IMPORTANT_**: We encourage contributions to the project from the community. We ask that you keep the following
guidelines in mind when planning your contribution.

  * Whether your contribution is for a bug fix or a feature request, **create an [Issue][create-issue]** and let us know
   what you are thinking.
  * **For bugs**, if you have already found a fix, feel free to submit a Pull Request referencing the Issue you created.
  * **For feature requests**, we want to improve upon the library incrementally, which means small changes at a time. In
   order ensure your PR can be reviewed in a timely manner, please keep PRs small, e.g. <10 files and <500 lines
   changed. If you think this is unrealistic, then mention that within the issue and we can discuss it.

Once you're ready to contribute code back to this repo, start with these steps:

  * Fork the appropriate sub-projects that are affected by your change
  * Clone the fork to `$GOPATH/src/github.com/intelsdi-x/`

	```
	$ git clone https://github.com/<yourGithubID>/<project>.git
	```

  * Create a topic branch for your change and checkout that branch

    ```
    $ git checkout -b some-topic-branch
    ```
    
  * Make your changes and run the test suite if one is provided (see below)
  * Commit your changes and push them to your fork
  * Open a pull request for the appropriate project
  * Contributors will review your pull request, suggest changes, and merge it when it’s ready and/or offer feedback
  * To report a bug or issue, please open a new issue against this repository

If you have questions feel free to contact the [snap maintainers][snap-maintainers].

### Vagrant development environment
Included in this repo is a Vagrant development environment that provisions a single-node Mesos cluster for developing
and testing this plugin. To get started, you'll need to have:

  * VirtualBox: <https://www.virtualbox.org/wiki/Downloads>
  * Vagrant: <https://www.vagrantup.com/downloads.html>

Next, assuming that your current working directory is the root of this Git repository, simply run the following command:

```
$ vagrant up
```

The provisioning script will install Mesos, ZooKeeper, Go, and Snap. You can then connect to Mesos at the following
URLs:

  * Master: <http://10.180.10.180:5050>
  * Agent: <http://10.180.10.180:5051>

In order to launch a task on the cluster that you can then observe with Snap, try running the following command in a
separate terminal window / SSH session:

```
$ mesos execute --master="$(mesos resolve zk://10.180.10.180:2181/mesos)" \
    --name="PythonHttpServer" --command="python -m SimpleHTTPServer 9000"
```

## Contributing Examples
The most immediately helpful way you can benefit this project is by cloning the repository, adding some further examples
and submitting a pull request.

Have you written a blog post about how you use [snap](http://github.com/intelsdi-x/snap) and/or this plugin? Send it to
us!

## Contributing Elsewhere
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at
<http://github.com/intelsdi-x/snap>.

## Thank You
And **thank you!** Your contribution, through code and participation, is incredibly important to us.


[create-issue]: https://github.com/intelsdi-x/snap-plugin-collector-mesos/issues
[snap-maintainers]: https://github.com/intelsdi-x/snap/blob/master/README.md#maintainers
