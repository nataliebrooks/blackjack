// MATH 221 FALL 2016
//
// CLIENT for a BLACKJACK CARD-PLAYING SERVICE
//

package main

import (
	"cs221"
	"fmt"
	"strconv"
	"os"
)

//
// main():
//
// Implements the command
//
//    go run player.go <host> <port>
//
// which plays a hand of blackjack with a 'dealer' service,
// located on the machine named '<host>', listening on the
// TCP/IP port numbered '<port>'.
// 
// For example, the command
//
//    go run player.go localhost 3000
//
// will connect to a dealer on the same machine, run with that
// same set of arguments, at a port numbered 3000. The command
//
//    go run player.go ravioli.reed.edu 3001
//
// will connect to a dealer running on ravioli at port 3001.
//
// The port should be at least 1024 and should match the port
// of some running dealer.
//
// This asks for the user's name, then asks them to 'H' (hit,
// request another card) or 'S' (stand, complete the hand)
// contining these rounds until the user stands.  It then
// reports the result of the game, played against the 
// dealer.
//
// The code relies on the 'cs221' package to establish the
// TCP/IP network connection with the server, then to exchange
// messages from the server.  These are sent and received over
// two Go channels (of type chan string) cout and cin (the
// "out to server" channel and the "in from the server" channel)
// provided by the 'MakeConnection' function.
//
// MakeConnection sets up a goroutine "proxy" that performs the
// actual network communications.
//
// To send a message, send a series of lines, for example
//
//    cout <- "Hello.\n"
//    cout <- "It's me.\n"
//    cout <- "\n"
//
// ending with an empty line.  
// 
// Then receive a message by a call to
//
//    message := <-cin
//
// This 'message' will be a single string of the form:
//
//    "I was wondering\nif after all these years\nyou'd like to meet
//To go over everything.\n\n"
//
// which corresponds to a send of the lines
//
//    I was wondering
//    if after all these years
//    you'd like to meet
//    To go over everything.
//
// followed by a blank line.
// 
// So what we are seeing here is that we use the 'out' channel
// to send a series of communications, ending with a blank line.
// And then what we get is a single communication response from
// the 'in' channel, packaging lines as ome string.
//
// The individual lines can be obtained with the helper
// function
//
//     cs221.Lines(message)
//
// which gives back an arrat/list of strings, for example
//
//    ["I was wondering","if after all these years",
//     "you'd like to meet", "To go over everything."]
//
// Most communications, you'll find, can just be single-line 
// message (followed by a blank line). You can get that single
// line's content instead with the helper function
//
//     cs221.Lines(message)
//
// The comments below describe the 'player/dealer' protocol 
// more carefully.
//
// The client can close a session by sending "" over cout.
//
// Both the client and server can use fmt.Scanf to pick apart
// the contents of a line of text, extracting the string's
// contents.

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Error: too few command-line arguments.\n")
		fmt.Printf("usage: go run player.go <host> <port>\n")
		os.Exit(0)
	}

	// Get the hostname from the command line.
	hostname := os.Args[1]
	// Get the port number.
	port,_ := strconv.Atoi(os.Args[2])

	// Make a connection, getting "proxy" channels cout/cin.
	cout, cin, e := cs221.MakeConnection(hostname, port, "player")
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}

	// Get the player's information. Send it to the server.
	fmt.Print("Enter your name: ")
	var name string	
	fmt.Scanln(&name)	
	cout <- name + "\n"
	cout <- "\n"

	// Play a hand of blackjack.  Note that this is a
	// "thin" client.  Most of the interaction involves
	// repeating what is sent from the server and 
	// forwarding responses from the user.
	//
	done := false
	for !done {
		//
		// Get a response from the server. Output it.
		reply := <-cin
		fmt.Println(reply)

		// See if the gameplay is over.
		lines := cs221.Lines(reply)

		// If the game is over, the last line of
		// the dealer's communication will be
		// "GAME OVER".
		//
		if lines[len(lines)-1] == "GAME OVER" {

			//
			// End the proxy's session by sending "".
			cout <- ""
			done = true

		} else {

			// Get a response from the user.
			var message string	
			fmt.Scanln(&message)

			// Forward it to the server.
			cout <- message + "\n"
			cout <- "\n"

			// For blackjack, this message ought to be
			// either "H" or "S".
		}
	}
}
