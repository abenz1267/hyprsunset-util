# hyprsunset-util

Utility for controlling display temperature with hyprsunset.

## Usage

```
#example
hyprsunset-util -duration=10 --sunset=17:00 --sunrise=07:00
```

By default, sunset and sunrise will be queried from the internet.

## Options

```
  -def string
    	default temperature. default: 6500 (default "6500")
  -disable
    	disable
  -duration int
    	transition duration time in minutes. default: 0
  -enable
    	enable
  -sunrise string
    	sunrise time (format: HH:MM). default: auto
  -sunset string
    	sunset time (format: HH:MM). default: auto
  -temp string
    	desired temperature. default: 3000 (default "3000")
```
