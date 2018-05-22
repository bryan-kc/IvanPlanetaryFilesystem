# Ivan Planetary Filesystem

CPSC 416 Project 2.

http://www.cs.ubc.ca/~bestchai/teaching/cs416_2017w2/project2/index.html

Created with ❤️ by: Jonathan Budiardjo @jopika, Jinny Byun @jinichu, Bryan Chiu, Tristan Rice @d4l3k, Bea Subion @biancasubion

## Problem Statement (Introduction)

The original implementation of the Interplanetary File System (IPFS) was
meant to provide an infrastructure for a distributed file system for the
dissemination of versioned data.
The system supports file chunk versioning using cryptographic hashes, like Git,
and uses a BitTorrent-like dissemination protocol, allowing the system to work
without a centralized server.
Although IPFS supports node verification through public keys, data
transferred between intermediate nodes from a source to a destination
have unencrypted access to file contents.
This allows for potentially malicious third parties to easily censor
documents while enroute to the end nodes.

Thus, we replicated a subset of IPFS features while also extending IPFS
to be able to encrypt data between the source and destination such that
intermediate nodes are not able to peek at the file contents being transferred
through the network. We have also implemented a publishing and subscription
system that allows for end users to subscribe to a reference,
and listen to messages that are published to the reference on a channel.

You can read more about the design of the project here: [DESIGN.md](docs/DESIGN.md)

## Instructions

- Git clone or copy the project code into a folder in your $GOPATH/ .

- Install [protobuf](https://github.com/google/protobuf/releases) (Google's Protocol
Buffer compiler).

- Run `make deps` inside of the project to install all of the go dependencies.

- Run `make` inside of the project to build the command line client and server binaries and run the
tests.

### Running
- To launch a node you can run `./proj2 -bind :8181`, second node should be run as
`./proj2 -bind :8282 -bootstrap localhost:8181 -path tmp/node2`.

- Command line interface can be run via `./ipfs localhost:8181` to connect to the
first node. Run `help` to find out more about those commands.

- There's a web interface also available at the node address (ex
https://localhost:8181) which runs on every node.


## CLI Commands

The Ivan Planetary File System’s command line client has 11 commands in
total for the user to interact with an IPFS node.
The command line client can be run by calling ​`go build ipfs.go`
in the ​`app`​folder and then running `./ipfs <node_addr>`.​

- `get <document_access_id>`

Fetches a document with this access ID and returns the contents of the document.
The access ID is in the format of ​`document_id:access_key`.
Since all documents are encrypted, the access key is used to decrypt the
document so that the contents can be retrieved. If `access_id` belongs to
a directory (a document with children), it will return a list of all the
children documents’ names and their access IDs instead.

- `add <path/to/file>`

Adds a local document to the IPFS and returns the access ID of the document,
in the format of `document_id:access_key`.

- `add -r <path/to/directory>`

Adds a local directory to the IPFS and returns the access ID of the document,
in the format of `document_id:access_key`.

- `add -c <documents>`

Creates a parent document to a list of existing documents in the IPFS and
returns the access ID of the document.
Documents must be in the format of ​`file_name,access_id`​and each document must be semicolon separated.
For example:
`index.html,document1_id:access_key1;foo.html,document2_id:access_key2`

- `peers list`

Lists all of the peer addresses of this node.

- `peers add <node_address>`

Add a peer to this node using the given address.

- `reference get <reference_access_id>`

Fetches what this reference points to (either a document or another reference)
and returns its record, in the format of ​`document@document_id:access_key`​​
or `reference@reference_id:access_key`.

- `reference add <record> <path/to/priv_key>`

Adds a reference to the IPFS (or updates an existing reference) and
returns the access ID.
The record is in the format of​`document@document_id:access_key`​or `reference@reference_id:access_key`.
The returned access ID is in the format of `reference_id:access_key`.

- `publish <message> <path/to/priv_key>`

Publishes a message to a reference on a channel. Nodes subscribed to this reference will then see the message.

- `subscribe <reference_id>`

Subscribes to an existing reference and listens for messages on a channel.
If a message is published to this reference, they will be seen on this channel.

- `quit`

Disconnects from this node and exits the Ivan Planetary File System.
