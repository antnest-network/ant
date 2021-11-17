# ant

## What is Ant?

ANT is a global, versioned, peer-to-peer filesystem. 

### System Requirements

ANT can run on most Linux, macOS, and Windows systems. We recommend running it on a machine with at least 2 GB of RAM and 2 CPU cores (ant is highly parallel). On systems with less memory, it may not be completely stable.


### Build from Source

ant's build system requires Go 1.15.2 and some standard POSIX build tools:

* GNU make
* Git
* GCC (or some other go compatible C Compiler) (optional)

To build without GCC, build with `CGO_ENABLED=0` (e.g., `make build CGO_ENABLED=0`).

#### Install Go

The build process for ant requires Go 1.15.2 or higher. If you don't have it: [Download Go 1.15+](https://golang.org/dl/).

You'll need to add Go's bin directories to your `$PATH` environment variable e.g., by adding these lines to your `/etc/profile` (for a system-wide installation) or `$HOME/.profile`:

```
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$GOPATH/bin
```

(If you run into trouble, see the [Go install instructions](https://golang.org/doc/install)).

#### Download and Compile ANT

```
$ git clone https://github.com/antnest-network/ant.git

$ cd ant
$ make build
$ make install
```

Alternatively, you can run `make build` to build the ant binary (storing it in `cmd/ant/ant`) without installing it.

**NOTE:** If you get an error along the lines of "fatal error: stdlib.h: No such file or directory", you're missing a C compiler. Either re-run `make` with `CGO_ENABLED=0` or install GCC.

##### Cross Compiling

Compiling for a different platform is as simple as running:

```
make build GOOS=myTargetOS GOARCH=myTargetArchitecture
```

### Usage

```
  ant - Global p2p merkle-dag filesystem.

  ant [<flags>] <command> [<arg>] ...

SUBCOMMANDS
  BASIC COMMANDS
    init          Initialize local ANT configuration

  ADVANCED COMMANDS
    daemon        Start a long-running daemon process

  NETWORK COMMANDS
    id            Show info about ANT peers
    bootstrap     Add or remove bootstrap peers
    swarm         Manage connections to the p2p network
    ping          Measure the latency of a connection


  TOOL COMMANDS
    version       Show ANT version information
    commands      List all available commands
    log           Manage and show logs of running daemon
    
  MINING COMMANDS
    cheque        Interact with cheques
    wallet        Interact with the wallet
    
  Use 'ant <command> --help' to learn more about each command.

  ant uses a repository in the local file system. By default, the repo is located at
  ~/.ant. To change the repo location, set the $ANT_PATH environment variable:

    export ANT_PATH=/path/to/antrepo
```
