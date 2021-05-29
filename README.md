## Introduction
Fire is a TCP game server for [Flame](https://github.com/Isangeles/flame) RPG engine, which enables multiple users
to connect and play together.

The server serves as a simple interface that handles connected users and offers them a set of requests to control
their characters and interact with the game world hosted on the server.

Communication between client and server is realized through a JSON request/response system.

Currently in a early development stage.
## Build
Get sources from git:
```
go get -u github.com/isangeles/fire
```
Install server executable:
```
go install github.com/isangeles/fire@latest
```
Or with GOPATH mode simply:
```
go install github.com/isangeles/fire
```
After that, the server executable will be placed in your GOBIN directory(eg. ~/go/bin).
## Run
Before starting server executable configure host address, port, and ID of Flame module in `.fire` file placed in the executable directory(create if it doesn't already exist):
```
host:[host]
port:[port]
module:[module ID]
```
Without address and host configuration, the server will use `localhost:8000` by default.

Module ID is the name of the module directory placed in `data/modules` in the server executable directory.

Flame modules are available to download [here](http://flame.isangeles.pl/mods).

Run server:
```
./fire
```
After this, the server is ready to handle incoming connections from the client programs.
## Clients
Any program able to send data through a TCP connection could serve as a Fire client.

For example, you can use [Ncat](https://nmap.org/ncat) utility to receive responses and make requests to the server.
Of course, interpreting server responses with just Ncat will be difficult, that's why some kind of specialized program is recommended.
[Burn Shell](https://github.com/isangeles/burnsh) and [Mural](https://github.com/isangeles/mural) are examples of interfaces that enable the user to play the game hosted on the Fire server.

Client programs use JSON based interface to communicate with the server via a set of requests and responses.

For each new connection, server sends a logon response to client, which is JSON in following format:
```
{"logon":true}
```
First thing that server client need to do, is to send valid login request in following format:
```
{"login":[{"id":"[user ID]","pass":"[user password]"}]}
```
After successful login, server will answer:
```
{"logon":false}
```
Each logged client is constantly updated with the current state of a Flame module through an update response.

Logged clients can use different JSON requests to modify their characters and interact with others on the server.

Check documentation for a detailed description of all available requests and server responses.
## Users
Users are stored in the `data/users` directory in the server executable directory.

Each user has its own directory with `.user` configuration file.

The name of a user directory is used as a unique user ID.

The user configuration file contains a password and list of flags for game characters controlled by the user.

Example user configuration:
```
pass:asd123!
char-flags:userAsdChar1;userAsdChar2
```
Check documentation for a detailed description of the user directory.
## Configuration
Server configuration is stored in `.fire` file placed in the server executable directory.
### Configuration values:
```
host:[host name]
```
Value for server host address, `localhost` by default.
```
port:[host port]
```
Value for server host port number, `8000` by default.
```
module:[module ID]
```
Name of the directory with a Flame module for the game hosted on the server.

The module should be placed in the `data/modules` directory in the server executable directory.
```
update-break:[duration in milliseconds]
```
Duration of update break after each game update in milliseconds.

If not set, the default value is 16 milliseconds(which should match the client's GUI running on 60 FPS).
```
action-min-range:[range value]
```
The minimum range required for game objects to interact with each other.
## Documentation
Source code documentation could be easily browsed with the `go doc` command.

Besides that `doc` directory contains documentation pages for request/response system and server files.

Documentation pages are in Troff format and could be easily displayed with `man` command.

For example to display documentation page for login request:
```
man doc/request/login
```
## Contributing
You are welcome to contribute to project development.

If you looking for things to do, then check the TODO file or contact maintainer(dev@isangeles.pl).

When you find something to do, create a new branch for your feature.
After you finish, open a pull request to merge your changes with master branch.
## Contact
* Isangeles <<dev@isangeles.pl>>
## License
Copyright (C) 2020-2021 Dariusz Sikora <<dev@isangeles.pl>>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
