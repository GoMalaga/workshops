// Block 3: a ticket office on the console. It reads one command per line from
// stdin, runs it against the Store and prints one line per command.
//
//	reserve <product> <qty>   -> "r1 Held": the new reservation's ID and status
//	confirm <reservation>     -> "ok"
//	cancel <reservation>      -> "ok"
//	stock <product>           -> "7": units still available
//	quit                      -> exits (so does the end of the input)
//
// Any failure prints "error: <message>" and the program keeps reading.
// Try it:  printf 'reserve p1 3\nstock p1\n' | go run ./cmd/taquilla
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"reservations"
)

func main() {
	store := reservations.NewStore()
	store.AddProduct(&reservations.Product{ID: "p1", Name: "Ticket", Stock: 10})
	if err := run(os.Stdin, os.Stdout, store); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run is the loop: read a line, split it into words, execute, print.
// It takes an io.Reader and an io.Writer instead of os.Stdin and os.Stdout so
// a test can feed it a scripted session (see main_test.go).
func run(in io.Reader, out io.Writer, store *reservations.Store) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		if len(words) == 0 {
			continue
		}
		if words[0] == "quit" {
			return nil
		}
		line, err := execute(store, words[0], words[1:])
		if err != nil {
			fmt.Fprintln(out, "error:", err)
			continue
		}
		fmt.Fprintln(out, line)
	}
	return scanner.Err()
}

// execute runs one command and returns the line to print.
// Errors from the Store come back as they are: the loop prints them.
func execute(store *reservations.Store, cmd string, args []string) (string, error) {
	switch cmd {
	case "reserve":
		if len(args) != 2 {
			return "", errors.New("usage: reserve <product> <qty>")
		}
		qty, err := strconv.Atoi(args[1])
		if err != nil {
			return "", fmt.Errorf("%q is not a number", args[1])
		}
		r, err := store.Reserve(args[0], qty)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s %v", r.ID, r.Status), nil
	case "confirm":
		if len(args) != 1 {
			return "", errors.New("usage: confirm <reservation>")
		}
		if err := store.Confirm(args[0]); err != nil {
			return "", err
		}
		return "ok", nil
	case "cancel":
		if len(args) != 1 {
			return "", errors.New("usage: cancel <reservation>")
		}
		if err := store.Cancel(args[0]); err != nil {
			return "", err
		}
		return "ok", nil
	case "stock":
		if len(args) != 1 {
			return "", errors.New("usage: stock <product>")
		}
		n, err := store.Stock(args[0])
		if err != nil {
			return "", err
		}
		return strconv.Itoa(n), nil
	default:
		return "", fmt.Errorf("unknown command %q", cmd)
	}
}
