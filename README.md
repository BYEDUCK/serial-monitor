Created as an alternative to Arduino IDE serial monitor/plotter for terminal enjoyers.

# Usage

```sh
go build .
./serial-monitor --baud 115200 --mode PLOT
```
```sh
go run main.go --baud 115200
```

## Help

```sh
./serial-monitor --help
```
```sh
go run main.go --help
```

# Controls

|   key   |                 action                     |
|---------|--------------------------------------------|
|**i**    |enter input mode                            |
|**ESC**  |exit program/input mode                     |
|**p**    |pause/unpause (close/open serial connection)|
|**m**    |change gui mode TEXT<-->PLOT                |
|**z**    |zoom in/out (enter/exit full screen)        |
|**c**    |clear message buffer                        |
|**j**    |scroll half page down[^1]                   |
|**k**    |scroll half page up[^1]                     |
|**b**    |scroll to bottom[^1]                        |
|**t**    |scroll to top[^1]                           |
|**f**    |enter/exit follow mode[^1]                  |
|**s**    |show/hide timestamps[^1]                    |
|**h**    |enter/exit hex mode[^1]                     |



[^1]: Only in **TEXT** gui mode