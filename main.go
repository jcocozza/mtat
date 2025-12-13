package main

import (
	"flag"
	"fmt"
	"os"
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
	fmt.Fprintln(os.Stderr, "  feed        return feed url for station id")
	fmt.Fprintln(os.Stderr, "  lines       list subway lines and corresponding feed url")
	fmt.Fprintln(os.Stderr, "  stations    list station ids and their common names")
	fmt.Fprintln(os.Stderr, "  serve       serve arrival times on an http server")
	fmt.Fprintln(os.Stderr, "  arrivals    get arrival times for a station")
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
	case "feed":
		feedCmd := flag.NewFlagSet("feed", flag.ExitOnError)
		err := feedCmd.Parse(args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		feedArgs := feedCmd.Args()
		if len(feedArgs) == 0 {
			fmt.Fprintln(os.Stderr, "station id required")
			os.Exit(1)
		}
		stationId := feedArgs[0]
		_, lineToStation, err := GetMappings()
		if err != nil {
			panic(err)
		}

		feed, ok := lineToStation[stationId]
		if !ok {
			fmt.Fprintf(os.Stderr, "invalid station id: %s\n", stationId)
			os.Exit(1)
		}
		_, err = fmt.Fprintln(os.Stdout, feed)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	case "stations":
		stationsCmd := flag.NewFlagSet("stations", flag.ExitOnError)
		delimiter := stationsCmd.String("d", "\t", "delimiter")
		err := stationsCmd.Parse(args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		stations, err := ReadStations()
		if err != nil {
			panic(err)
		}
		for _, station := range stations {
			_, err := fmt.Fprintf(os.Stdout, "%s%s%s\n", station.Id, *delimiter, station.Name)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
	case "lines":
		linesCmd := flag.NewFlagSet("lines", flag.ExitOnError)
		delimiter := linesCmd.String("d", "\t", "delimiter")
		err := linesCmd.Parse(args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		for subway, url := range SubwayFeedMap {
			_, err := fmt.Fprintf(os.Stdout, "%s%s%s\n", subway, *delimiter, url)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
	case "serve":
		saCmd := flag.NewFlagSet("serve", flag.ExitOnError)
		port := saCmd.Int("port", 8080, "port to run the server on")
		err := saCmd.Parse(args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		err = Serve(*port)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	case "arrivals":
		arrivalsCmd := flag.NewFlagSet("arrivals", flag.ExitOnError)
		stationDirection := arrivalsCmd.String("d", "", "direction (N/S; default BOTH)")
		arrivalsCmd.Usage = func() {
			fmt.Fprintf(os.Stderr, "%s arrivals [OPTIONS] [station id]\n", os.Args[0])
			fmt.Fprintln(os.Stderr, "get upcoming arrivals by station")
			arrivalsCmd.PrintDefaults()
		}
		err := arrivalsCmd.Parse(args[1:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		stationArgs := arrivalsCmd.Args()
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

		var arrivals []Arrival
		for _, stopId := range stopIDs {
			a, err := GetFutureArrivals(stopId)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] %v", err)
				os.Exit(1)
			}
			arrivals = append(arrivals, a...)
		}

		fmt.Println("arrival time, route id, stop id")
		for _, arrival := range arrivals {
			//fmt.Println(arrival.ArrivalTime.Format("15:04:05"))
			fmt.Println(arrival.ArrivalTime, arrival.RouteID, arrival.StopID)
		}
	}
}
