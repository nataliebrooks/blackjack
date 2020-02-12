// MATH 221 Fall 2016
//
// CARD-PLAYING SERVICE
//
// Below is my solution to Homework 9 Exercise 2, modified
// to use network code.  It provides a 'dealer' service,
// a server that is willing to accept connections from
// 'player' clients, deal them a hand of blackjack, and
// then ask them to play out the hand.
//
// The code is nearly the same as the code at 
//
//   http://jimfix.github.io/math221/wk13/hw10ex2.go
// 
// Except 'Printf' and 'Println' occurrences are replaced 
// with code that sends a series of lines (represented as
// strings ending in the character '\n') over a 'cout'
// channel, and 'Scanln' is replaced by code that receives 
// a client's "hit or stand" play over a 'cin' channel.
//
// We do this because 'cout' and 'cin' are channels that
// communicate with a "proxy" go routine, one that forwards
// information over the network to/from a connected 
// client (one that connects to us on behalf of a user).
// This makes 'main' conceptually a bit more complicated.  
// Using a MATH221 package that supports 'client/server' 
// connections, the 'dealer' server creates a TCP/IP port 
// for listening to connecting clients/players. With each 
// player that connects, it runs 'handleGame' over two
// channels that are used to represent the connection.
//
// To see how this cout/cin communication mechanism works
// You might want to take a look at the 'handlePlay'
// function, particularly its use of 'cout' and 'cin'
// within the helper functions 'reportHand' and 'onePlay'.  
// You should also take a careful look at the corresponding 
// use of 'cout' and 'cin' in the code for 'player.go', 
// particularly in its 'main' loop. The comments in that 
// file explain the way a series of lines are packaged
// up and sent as a single string.
//
// To run this code, you type something like
//
//    go run dealer.go <host> <port>
//
// which allows 'player' clients to connect and play a
// hand of blackjack. The <host> needs to be the name of
// the machine on which you are running this Go dealer 
// program and the <port> number has to be some integer 
// you choose, 1024 or larger, that is not being used by 
// other services. For example, if you run the command
//
//    go run dealer.go ravioli.reed.edu 3000
//
// when you are logged onto 'ravioli' then this will bring
// up a dealer that listens for connections from player 
// clients on the port numbered 3000. 
//
// You can also allow only local machine connections by
// instead typing something like
//
//    go run dealer.go localhost 3000
//
// This will only allow player.go clients on the same 
// machine to connect, and they will have to connect
// using a <host> of "localhost", too.
//
package main

import (
    "fmt"
    "math/rand"
    "strconv"
    "time"
    "os"
    "cs221"

)

// Global channel shared by all connection handlers.
// Used to report game results.
type report struct {
    player string
    result string
}
var creport chan report = make(chan report)


// BACKGROUND_IS_BRIGHT : bool
//
// This flag should be true if the Terminal background
// is bright; false if it is dark.
//
var BACKGROUND_IS_BRIGHT bool = false

// number : int -> int
//
// Converts a card's code (0-51) into it's number,
// from 1 (for Ace) to 13 (for King).
//
func number(c int) int {
    return (c%13 + 1)
}

// rank : int -> string
//
// Converts a card's code (0-51) into it's rank,
// a string that is A, 2, etc., 10, J, Q, K.
//
// See makeDeck below for details.
//
func rank(c int) string {
    var face = map[int]string{
        1:  "A",
        11: "J",
        12: "Q",
        13: "K",
    }
    var r int = number(c)
    if r >= 2 && r <= 10 {
        return strconv.Itoa(r)
    } else {
        return face[r]
    }
}

// value : int -> int
//
// Converts a card's code (0-51) into it's score
// value from 1 to 10.  An Ace is valued at 1.
//
func value(c int) int {
    var r int = number(c)
    if r < 10 {
        return r
    } else {
        return 10
    }
}

// suit : int -> string
//
// Converts a card's code (0-51) into it's suit,
// a unicode character for spades, clubs, hearts,
// or diamonds.
//
// See https://en.wikipedia.org/wiki/
//             Playing_cards_in_Unicode
//
// See makeDeck below for details.
//
func suit(c int) string {

    var suits map[int]string

    if BACKGROUND_IS_BRIGHT {
        suits = map[int]string{
            0: "\u2660", // spades
            1: "\u2663", // clubs
            2: "\u2661", // hearts
            3: "\u2662", // diamonds
        }
    } else {
        suits = map[int]string{
            0: "\u2664", // spades
            1: "\u2667", // clubs
            2: "\u2665", // hearts
            3: "\u2666", // diamonds
        }
    }
    var s int = c % 4
    return suits[s]
}

// card : int -> string
//
// Given a playing card's code (0-51), returns a
// string depicting that card.
//
func card(c int) string {
    return rank(c) + suit(c)
}

// makeDeck: . -> []int
//
// Creates a 52 card deck, a random permutation of
// the card code values from 0 to 51.
//
// Regarding the code, each group of 13 codes is in
// the same suit and then, within that range, they
// are ordered from Ace up to King.  The suits' code
// ranges are as follows:
//
//    0-12  spades
//    13-25 clubs
//    26-38 hearts
//    39-51 diamonds
//
func makeDeck() []int {
    var deck []int = make([]int, 52)
    for position, _ := range deck {
        deck[position] = position
    }
    rand.Seed(time.Now().UnixNano())
    for position, _ := range deck {
        var swapWith int = rand.Intn(52 - position) + position
        if swapWith != position {
            deck[position], deck[swapWith] = deck[swapWith], deck[position]
        }
        position = position + 1
    }
    return deck
}

// tallyHand : []int -> int
//
// Evaluates an array of playing cards as a hand
// according to BlackJack scoring.  Aces are worth
// 11 unless that makes the hand go over 21, in
// which case they can count as 1 instead.
//
func tallyHand(hand []int) int {
    var hasAces bool = false // Keep track of whet
    var total int = 0
    for _, c := range hand {
        if rank(c) == "A" {
            hasAces = true
        }
        total = total + value(c)
    }
    if hasAces && total <= 11 {
        total = total + 10
    }
    return total
}

// reportHand : (chan<- string) x string x []int x bool -> .
//
// Given a player name, their hand of cards, and
// a "show" flag, reports that player's hand to
// the program user.  It also reports the player's
// BlackJack tally.
//
// In some cases, a BlackJack program may not want
// to reveal the player's first card and their
// tally.  In that case, showFirst should be set
// to false.
//
// The report is sent as a line of communication 
// over the channel named 'cout'.
//
func reportHand(c chan<- string, name string, 
                hand []int, showFirst bool) {

    // Output the name.
    c <- fmt.Sprintf(name + ": \t")

    // Output the tally.
    if showFirst {
        c <- fmt.Sprintf("WORTH: %2d\tHAND: ", tallyHand(hand))
    } else {
        // ...unless it should be hidden.
        c <- fmt.Sprintf("WORTH: ??\tHAND: ")
    }

    // Output the hand of cards, perhaps hiding the first.
    var first bool = true
    for _, cd := range hand {
        // Maybe hide the first card, but show the rest after.
        if first && !showFirst {
            c <- fmt.Sprintf("?? ")
        } else {
            c <- fmt.Sprintf(card(cd) + " ")
        }
        first = false // We've shown at least one.
    }

    // End this line of communication.
    c <- "\n"
}

// onePlay : . -> bool
//
// Determines whether the player would like to obtain
// another card, i.e., would they like to be "hit" or
// would they like to "stand"?
//
// Returns true if they want to continue play (hit to
// obtain a card), and false if they are done (want to
// stand).
//
// The interactions if performed by communicating over
// the pair of channels named 'cout' and 'cin'. The 
// first is used to send the prompt to the user. The
// second is used to read the reply from the user.
//
func onePlay(cout chan<- string, cin <-chan string) bool {

    // Requests an 'H' or an 'S'.
    cout <- fmt.Sprintf("Would you like to hit[H]? or stand[S]?\n")
    // Terminates message with a blank line.
    cout <- "\n"

    // Gets the reply, presumably 'H' or 'S'.
    reply := <-cin
    // Reads only the first line for 'H' or 'S'.
    s := cs221.HeadLine(reply)

    // Return true if they want another card (a hit).
    return (s == "H" || s == "h")
}

// handleGame
//
// This plays a game of one-player BlackJack using one
// shuffled deck of cards from a standard 52 card deck.
//
// After the first four cards are dealt, two to the
// dealer and two to the player, players choose to
// obtain more cards (get "hit") until they go over
// 21 (a "bust") or are satisfied with their hand.
//
// The player goes first.  The dealer then plays a
// fixed strategy to try to obtain a hand over 16.
//
// The dealer concedes, does not hit, when the player
// wins right away with 21.
//
// Communication to the user is via a series of messages
// sent over the channel 'cout' and read from the 
// channel 'cin'. 
//
func handleGame(cout chan<- string, cin <-chan string, info interface{}) {

    // Play starts with the player sending their name.
    clientname := cs221.HeadLine(<-cin)

    fmt.Println("Player '"+clientname+"' has connected.")

    //
    // Set up the deck and the two players' hands.
    var cards []int = makeDeck()
    var top int = 0 // The index of the first undealt card.
    // It is incremented with each card dealt.
    var dealer []int = make([]int, 0)
    var player []int = make([]int, 0)

    //
    // Deal the first two cards of each hand.
    player = append(player, cards[top])
    top++
    dealer = append(dealer, cards[top])
    top++
    player = append(player, cards[top])
    top++
    dealer = append(dealer, cards[top])
    top++
    fmt.Println("Player '"+clientname+"' has hand.")
    reportHand(cout, clientname, player, true)
    reportHand(cout, "house", dealer, false)
    fmt.Println("Player '"+clientname+"' is thinking.")

    //
    // Perform the game play of the player.
    for tallyHand(player) < 21 && onePlay(cout,cin) {
        player = append(player, cards[top])
        top++
        reportHand(cout, clientname, player, true)
        reportHand(cout, "house", dealer, false)
    }
    var playerTotal int = tallyHand(player)

    //
    // Maybe engage in dealer play.
    if playerTotal < 21 {

        //
        // The dealer follows a fixed strategy to try
        // and beat the player's score.
        cout <- fmt.Sprintf("Okay! You've held.\n")
        cout <- fmt.Sprintf("The dealer flips his first card.\n")
        reportHand(cout, "house", dealer, true)

        //
        // Keep hitting if under 17.
        for tallyHand(dealer) < 17 {
            dealer = append(dealer, cards[top])
            top++
            reportHand(cout, clientname, player, true)
            reportHand(cout, "house", dealer, true)
        }

    } else if playerTotal == 21 {

        //
        // The dealer is forced to stand if the player has 21.
        cout <- fmt.Sprintf("Twenty-one!\n")
        cout <- fmt.Sprintf("The dealer reveals his hand.\n")
        reportHand(cout, "house", dealer, true)
    }

    //
    // Report the results.
    var dealerTotal int = tallyHand(dealer)
    var outcome string
    if playerTotal > 21 {
        cout <- fmt.Sprintf("BUST! I'm sorry but you lose.\n")
        outcome = "lost"
    } else if dealerTotal > 21 {
        cout <- fmt.Sprintf("The dealer went over. You win!\n")
        outcome = "won"
    } else if playerTotal > dealerTotal {
        cout <- fmt.Sprintf("You win with the better total!\n")
        outcome = "won"
    } else if playerTotal < dealerTotal {
        cout <- fmt.Sprintf("Sorry. The dealer won.\n")
        outcome = "lost"
    } else {
        cout <- fmt.Sprintf("No winner. A push.\n")
        outcome = "tied"
    }
    cout <- "GAME OVER\n"
    cout <- "\n"
    fmt.Println("Player '"+clientname+"' has "+outcome+".")

	creport <- report{player:clientname, result:outcome}
}

//
// main():
//
// Implements the command
//
//    go run dealer.go <host> <port>
//
// Where <host> is the name of the machine you are on, and
// where <port> is the number (1024, or higher) clients use
// to connect to your service. For example
//
//    go run dealer.go ravioli.reed.edu 3000
//
// See the description at the top for details.
//
func main() {
    
	if len(os.Args) < 3 {
		fmt.Printf("Error: too few command-line arguments.\n")
		fmt.Printf("usage: go run dealer.go <host> <port>\n")
		os.Exit(0)
	}

	hostname := os.Args[1]
	port,_ := strconv.Atoi(os.Args[2])

    // Run a reporting go routine.
    go func() {
        for {
            r := <-creport 
            fmt.Printf("REPORT: %s %s\n",r.player,r.result) 
        }
    }()

    // Loop, accepting and handling client connections.
    e := cs221.HandleConnections(hostname, port, handleGame, "Dealer", nil)

    // This code will only be reached when the server
    // is shutting down, or when an error occurs.
    //
    if e != nil {
        fmt.Println(e.Error())
        os.Exit(0)
    }
}
