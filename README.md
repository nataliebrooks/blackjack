# blackjack

This is a simple concurrent program written in Go Programming Language that plays a game of blackjack.

Environment:

You must have GO installed.

This program relies on a CS221 package that helps with creating and connecting to the server.
It is supplied in the main folder of the repository, but must be copied to this location 
(or the location that GO is installed on your machine):

/usr/local/Cellar/go/1.7.3/libexec/src

Create a new folder titled "cs221" in the src folder, then copy the cs221.go file from this repository
into it.

Running the Program:

In order to run the game you must start the dealer server before a player enters:

To start dealer, run:  "go run dealer.go localhost 8888"

To start player, open another terminal and run: "go run player.go localhost 8888"
