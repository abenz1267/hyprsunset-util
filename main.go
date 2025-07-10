package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

const (
	DEFAULT    = "6500"
	NIGHT      = "3000"
	TRANSITION = 0
)

const (
	StateEnabled       = "enabled"
	StateDisabled      = "disabled"
	StateTransitioning = "transitioning"
)

type daylight struct {
	sunrise time.Time
	sunset  time.Time
}

func main() {
	var def, temp, sunriseCustom, sunsetCustom string
	var duration int
	var disable, enable bool

	flag.StringVar(&def, "def", DEFAULT, "default temperature. default: 6500")
	flag.StringVar(&temp, "temp", NIGHT, "desired temperature. default: 3000")
	flag.StringVar(&sunriseCustom, "sunrise", "", "sunrise time (format: HH:MM). default: auto")
	flag.StringVar(&sunsetCustom, "sunset", "", "sunset time (format: HH:MM). default: auto")
	flag.IntVar(&duration, "duration", TRANSITION, "transition duration time in minutes. default: 0")
	flag.BoolVar(&disable, "disable", false, "disable")
	flag.BoolVar(&enable, "enable", false, "enable")

	flag.Parse()

	if enable {
		changeNow(temp)
		return
	}

	if disable {
		changeNow(def)
		return
	}

	light := daylight{}

	if sunriseCustom == "" || sunsetCustom == "" {
		light = getTimeFromWeb()
	}

	if sunriseCustom != "" && sunsetCustom != "" {
		now := time.Now()

		sunrise, err := time.Parse("15:04", sunriseCustom)
		if err != nil {
			log.Fatal(err)
		}

		sunset, err := time.Parse("15:04", sunsetCustom)
		if err != nil {
			log.Fatal(err)
		}

		light.sunrise = time.Date(now.Year(), now.Month(), now.Day(), sunrise.Hour(), sunrise.Minute(), 0, 0, now.Location())
		light.sunset = time.Date(now.Year(), now.Month(), now.Day(), sunset.Hour(), sunset.Minute(), 0, 0, now.Location())
	}

	if light.sunrise.IsZero() || light.sunset.IsZero() {
		log.Fatal("sunrise or sunset time is not set")
	}

	changeNow(def)
	currentTemp := def

	state := StateDisabled

	for {
		if state == StateTransitioning {
			time.Sleep(time.Minute * 1)
			continue
		}

		now := time.Now()

		if now.After(light.sunrise) && now.Before(light.sunset) {
			if state == StateEnabled {
				state = StateTransitioning
				change(def, currentTemp, duration)
				currentTemp = def
				state = StateDisabled
				log.Println("transitioned to", currentTemp)
			}
		} else if now.After(light.sunset) {
			if state == StateDisabled {
				state = StateTransitioning
				change(temp, currentTemp, duration)
				currentTemp = temp
				state = StateEnabled
				log.Println("transitioned to", currentTemp)
			}
		}

		time.Sleep(time.Minute * 1)
	}
}

func change(target, current string, duration int) {
	log.Println("changing to", target)

	if duration == 0 {
		changeNow(target)
		return
	}

	t, _ := strconv.Atoi(target)
	c, _ := strconv.Atoi(current)

	step := (t - c) / duration

	s := ""

	if step >= 0 {
		s = fmt.Sprintf("+%d", step)
	} else {
		s = fmt.Sprintf("%d", step)
	}

	count := 0

	for count < duration {
		cmd := exec.Command("hyprctl", "hyprsunset", "temperature", s)

		log.Println(cmd.String())

		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			cmd.Wait()
		}()

		count++
		time.Sleep(time.Second * 5)
	}
}

func changeNow(temp string) {
	log.Println("setting to", temp)

	cmd := exec.Command("hyprctl", "hyprsunset", "temperature", temp)

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		cmd.Wait()
	}()
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func getLocation() Location {
	endpoint := "http://ip-api.com/json/?fields=lon,lat"

	resp, err := http.Get(endpoint)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var location Location
	err = json.Unmarshal(body, &location)
	if err != nil {
		log.Fatal(err)
	}

	return location
}

func getTimeFromWeb() daylight {
	location := getLocation()

	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%f&lng=%f&formatted=0", location.Lat, location.Lon), nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data struct {
		Results struct {
			Sunrise string `json:"sunrise"`
			Sunset  string `json:"sunset"`
		} `json:"results"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Fatal(err)
	}

	format := "2006-01-02T15:04:05-07:00"
	now := time.Now()
	l := now.Location()

	sunrise, err := time.Parse(format, data.Results.Sunrise)
	if err != nil {
		log.Fatal(err)
	}

	sunrise = sunrise.In(l)

	sunset, err := time.Parse(format, data.Results.Sunset)
	if err != nil {
		log.Fatal(err)
	}

	return daylight{sunrise, sunset}
}
