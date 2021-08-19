package main

import (
	"fmt"
	"time"

	"bufio"
	"log"
	"notifier/notify"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"reflect"
	"strconv"
	"sync"
	"syscall"
)

func worker(done chan bool, message string, signalReceivedChannel chan os.Signal, responseChannel chan interface{}, lib notify.API, requestPayload *notify.Request, url string) {
	fmt.Print("working...")

	// Perform a go call to Notify function of notify library.
	go lib.Notify(requestPayload, responseChannel, url)

	// Last line of user's input file.
	// Set done channel to true
	if message == "" {
		fmt.Println("done")
		done <- true
	}
}

func main() {

	// Define/initialize variables.
	var (
		signalReceivedChannel = make(chan os.Signal, 1)
		responseChannel       = make(chan interface{})
		done                  = make(chan bool)
		wg                    sync.WaitGroup
		interval              int
		response              []interface{}
	)

	// Define client's terminal application.
	//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~USAGE~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
	//	1. Without interval input (default interval 1s)
	//			./notifier notify "http://localhost:8080/notify" < messages.txt
	//	2. With interval input (interval 5s)
	//			./notifier notify "http://localhost:8080/notify" "5" < messages.txt
	//~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~//
	var cmdNotify = &cobra.Command{
		Use:  "notify",
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {

			// Configure signals for signalReceivedChannel to receive.
			signal.Notify(
				signalReceivedChannel,
				syscall.SIGINT, // kill -SIGINT XXXX or Ctrl-x
			)

			// Add 2 go routines to the waitgroup.
			wg.Add(2)

			// Go routine that listens for SIGINT in order for the program to perform a graceful shutdown.
			go func() {
				defer wg.Done()
				<-signalReceivedChannel
				log.Print("os interrupt received - gracefully shutting down...\n")
				close(signalReceivedChannel)
				close(responseChannel)
				os.Exit(0)
			}()

			go func() {
				// On return, notify the WaitGroup the job is done.
				defer wg.Done()

				// Read interval input from args. If not present or zero then it defaults to 1s.
				if len(args) < 2 {
					interval = 1
					log.Println("default interval")
					log.Println(interval)
				} else {
					interval, _ = strconv.Atoi(args[1])
					if interval < 1 {
						log.Println("default interval")
						log.Println(interval)
						interval = 1
					}
					log.Println("current interval")
					log.Println(interval)
				}

				// Create a new lib object so as to be able to use the notify library.
				lib := notify.New()

				// Create a scanner object so as to use it as a message reader in the following loop.
				scanner := bufio.NewScanner(os.Stdin)
				if err := scanner.Err(); err != nil {
					fmt.Println(err)
				}

				// Loop that runs every interval in seconds.
				for tick := range time.Tick(time.Duration(interval) * time.Second) {
					// Prints UTC time and date - just for logging purposes.
					fmt.Println(tick, time.Now())

					// Read the next message from user's input file given.
					scanner.Scan()
					requestPayload := &notify.Request{
						Message: scanner.Text(),
					}

					// Performs a go routine call to the worker.
					go worker(done, requestPayload.Message, signalReceivedChannel, responseChannel, lib, requestPayload, args[0])
					if scanner.Text() == "" {
						break
					}
				}
			}()

			go func() {
				// On return, notify the WaitGroup the job is done.
				defer wg.Done()

				log.Println("Channel filled up with data. Iterating...")

				for item := range responseChannel {
					log.Println("current item")
					log.Println(item)

					objType := reflect.TypeOf(item).Elem().String()
					if objType == "notify.Response" {
						emptyNotifyResponse := notify.Response{}
						if *item.(*notify.Response) == emptyNotifyResponse {
							log.Println("Aggregated response objects.")
							log.Println(response)
							break
						} else {
							notifyItem := item.(*notify.Response)
							log.Println("notifyItem")
							log.Println(notifyItem)
							response = append(response, notifyItem)
						}
					} else if objType == "error" {
						errorItem := item.(error)
						log.Println("errorItem")
						log.Println(errorItem)
						response = append(response, errorItem)
					} else {
						log.Println("Something went wrong. Here are the aggregated response objects.")
						log.Println(response)
						break
					}
				}
			}()

			// Block until we receive a notification from the worker on the channel.
			<-done

			// Block until the WaitGroup's go routines return and then close the channels.
			wg.Wait()
			close(signalReceivedChannel)
			close(responseChannel)
			return
		},
	}

	// Configure root cmd command and add cmdNotify subcommand created above.
	var rootCmd = &cobra.Command{Use: ""}
	rootCmd.AddCommand(cmdNotify)
	rootCmd.Execute()
}
