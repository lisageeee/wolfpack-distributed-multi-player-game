### To install necessary Pixel dependencies on initial install, run from project root:
`./setup.sh`

### To run Wolfpack
##### First, start the server (locally or remote)
  `cd server ; go run server.go`
  
##### Start the logic node
`cd logic ; go run logic.go`

or, with optional command line args:

`go run logic.go [other-node-listener-addr] [pixel-incoming-addr] [pixel-outgoing-addr]`

##### Finally, start the Pixel node
`cd pixel ; go run pixel.go`

or, with optional command line args:

`go run pixel.go [logic-node-addr] [local-listener-addr]`

*Note: for a logic / pixel node pair, the last two arguments to the command line should be the same*