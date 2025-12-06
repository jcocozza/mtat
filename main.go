package main

import (
	"fmt"
	"os"
	"flag"
)

const (
	url = "https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs"
)

func usage() {
	fmt.Fprintf(os.Stderr, "%s [COMMAND] [OPTIONS] [ARGS]\n", os.Args[0])
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "a cli tool for getting nyc mta subway info.")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "commands:")
	fmt.Fprintln(os.Stderr, "  arrivals    get arrival times for a station")
	fmt.Fprintln(os.Stderr, "  serve       serve arrival times on an http server")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "%s <command> -h for more info on a command\n", os.Args[0])
}


func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		usage()
		return
	}

	switch args[0] {
	case "serve":
		saCmd := flag.NewFlagSet("serve", flag.ExitOnError)	
		port := saCmd.Int("port", 8080, "port to run the server on")
		saCmd.Parse(args[1:])
		err := Serve(*port)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	case "arrivals":
		stationCmd := flag.NewFlagSet("arrivals", flag.ExitOnError)
		stationDirection := stationCmd.String("d", "", "direction (N/S; default BOTH)")
		stationCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "%s arrivals [OPTIONS] [station id]\n", os.Args[0])
			fmt.Fprintln(os.Stderr, "get upcoming arrivals by station")
			stationCmd.PrintDefaults()
		}

		stationCmd.Parse(args[1:])

		stationArgs := stationCmd.Args()
		if len(stationArgs) != 1 {
			fmt.Fprintln(os.Stderr, "missing station id")
			os.Exit(1)
		}

		station := stationArgs[0]

		var stopIDs []string
		if *stationDirection == "" { // if no direction, we include both
			stopIDs = []string{
				fmt.Sprintf("%s%s", station, "N"),
				fmt.Sprintf("%s%s", station, "S"),
			}
		} else {
			stopIDs = []string{fmt.Sprintf("%s%s", station, *stationDirection)}
		}
		arrivals, err := GetFutureArrivals(url, stopIDs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] %v", err)
			os.Exit(1)
		}

		for _, arrival := range arrivals {
			fmt.Println(arrival.ArrivalTime.Format("15:04:05"))
		}
	}
}
