# logg
[![Go Reference](https://pkg.go.dev/badge/github.com/pkgz/logg.svg)](https://pkg.go.dev/github.com/pkgz/logg)
[![Tests](https://img.shields.io/github/actions/workflow/status/pkgz/logg/go.yml?branch=master)](https://github.com/pkgz/logg/actions)
[![codecov](https://img.shields.io/codecov/c/gh/pkgz/logg)](https://codecov.io/gh/pkgz/logg)

![](https://serhiy.s3.eu-central-1.amazonaws.com/Github_repo/logg/v3_pretty.png?v1)  
![](https://serhiy.s3.eu-central-1.amazonaws.com/Github_repo/logg/v3_json.png)  
Better log experience in golang.

## Installation
```bash
go get github.com/pkgz/logg
```

## Features
- color logs
- different output formats (pretty/json)
- different levels (debug/info/warning/error/panic/fatal)
- structured key/value fields
- drop-in `log/slog` handler
- can be used with internal log library
- zero allocations

## Usage
There is two way how you can use this logger:

- install globally (this option will set own writer to default log);
- using a local logger instance;

### Global logger
In this mode, the logger will set own writer to log the library. So you can use the default log library for logging. But with better output.

```golang
package main

import (
	"log"
	"os"

	"github.com/pkgz/logg"
)

func main() {
	logg.NewGlobal(os.Stdout)

	log.Print("ERR some message to log")
}
```

### Local logger
In this case, you must define a log instance in each context (function). But in this case, you can define different settings for different scopes.

```golang
package main

import (
	"os"

	"github.com/pkgz/logg"
)

func main() {
	log := logg.New(os.Stdout)

	log.Print("ERR some message to log")
}
```


### Json log

```golang
package main

import (
	"os"

	"github.com/pkgz/logg"
)

func main() {
	log := logg.New(os.Stdout)
	log.SetFormat(logg.Json)

	log.Print("ERR some message to log")
}
```

### Level functions
Instead of embedding the level in the message string, you can call a level function directly. These are available both on a local logger instance and as package-level functions after `NewGlobal`:

```golang
logg.Info("starting up")
logg.Warnf("retrying in %d seconds", 5)
logg.Error("something went wrong")
```

Each level has a plain and a formatted (`f`) variant: `Debug`/`Debugf`, `Info`/`Infof`, `Warn`/`Warnf`, `Error`/`Errorf`, `Panic`/`Panicf`, `Fatal`/`Fatalf`.

`Panic`/`Panicf` log the message and then `panic` with it; `Fatal`/`Fatalf` log the message and then call `os.Exit(1)`.

### Structured fields
Attach key/value pairs with `With`. The returned `*Entry` exposes the same level methods, and the fields are rendered as `key=value` pairs in pretty mode and as JSON keys in json mode:

```golang
log := logg.New(os.Stdout)

log.With(logg.F("user", "bob"), logg.F("attempt", 3)).Warn("login failed")
// INF login failed user=bob attempt=3

log.SetFormat(logg.Json)
log.With(logg.F("user", "bob"), logg.F("attempt", 3)).Warn("login failed")
// {"level": "WRN", "message": "login failed", "user": "bob", "attempt": 3}
```

`Entry` is immutable; `entry.With(...)` returns a new entry, so a base entry can be shared and extended safely.

### slog handler
`logg` can act as a backend for the standard library's `log/slog`, so you get logg's pretty/json output through the standard structured API:

```golang
import "log/slog"

logger := slog.New(logg.New(os.Stdout).SlogHandler())
logger.Info("request handled", "method", "GET", "status", 200)

// or, as a one-liner:
logger := logg.NewSlog(os.Stdout)
```

The handler honors the logger's format, time flags, color and minimum level. slog groups are flattened into dotted keys (`req.id`). Caller/source is not emitted via the slog bridge, so leave the file flags off when using it.

### Settings
There are a few parameters which you can set:

- flags (define time and caller format. Using the format from internal log library)
- format (output log format. Pretty or Json)
- color (colorize output or not)

#### Levels
When a level is embedded in the message string it is recognized in any of these forms:

- Debug (`logg.DebugLevel`): `DBG | DEBUG | [DBG] | [DEBUG]`
- Info (`logg.InfoLevel`): `INF | INFO | [INF] | [INFO]`
- Warning (`logg.WarningLevel`): `WRN | WARN | [WRN] | [WARN]`
- Error (`logg.ErrorLevel`): `ERR | ERROR | [ERR] | [ERROR]`
- Panic (`logg.PanicLevel`): `PNC | PANIC | [PNC] | [PANIC]`
- Fatal (`logg.FatalLevel`): `FTL | FATAL | [FTL] | [FATAL]`

The `*Level` constants are what you pass to `MinLevel`, e.g. `logg.MinLevel(logg.WarningLevel)`.

#### API
| Function | Default | Description |
| --- | --- | --- |
| `SetWriter(io.Writer) ` | io.Discard | Set writer. |
| `SetFormat(logg.format) ` | Pretty | Set output format. Can be pretty or json. |
| `SetFlags(int) ` | int | Set time and caller flags. |
| `MinLevel(level) ` | InfoLevel | Minimum level for logs. Logs lower this level will be not writed. |
| `ToggleColor(bool) ` | true | Enable or disable output colorizing. |
| `DebugMode() ` | | Will enable a debug mode. Debug mode will add milliseconds to timestamp and log caller. |

## Benchmarks

```sh
BenchmarkLog_Print/long_message-8         	 1799361	       703 ns/op	     592 B/op	       2 allocs/op
BenchmarkLog_Print/short_message-8        	 2444176	       491 ns/op	      80 B/op	       2 allocs/op
BenchmarkLog_Print/short_level-8          	 2451745	       494 ns/op	      80 B/op	       2 allocs/op
BenchmarkLog_Print/long_level-8           	 2364008	       493 ns/op	      80 B/op	       2 allocs/op
BenchmarkLogg_Log_Print/long_message-8    	 1443676	       825 ns/op	     593 B/op	       2 allocs/op
BenchmarkLogg_Log_Print/short_message-8   	 1701645	       713 ns/op	      80 B/op	       2 allocs/op
BenchmarkLogg_Log_Print/short_level-8     	 1882834	       636 ns/op	      80 B/op	       2 allocs/op
BenchmarkLogg_Log_Print/long_level-8      	 1862190	       676 ns/op	      80 B/op	       2 allocs/op
BenchmarkLogg_Write_Pretty/long_message-8 	13835029	        90.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Pretty/short_message-8         	13607654	        86.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Pretty/short_level-8           	16418979	        73.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Pretty/long_level-8            	15362362	        76.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Json/long_message-8            	11209905	       107 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Json/short_message-8           	11571448	       105 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Json/short_level-8             	13749747	        87.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Write_Json/long_level-8              	13415761	        89.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkLogg_Print/long_message-8                 	 4293176	       311 ns/op	    1170 B/op	       3 allocs/op
BenchmarkLogg_Print/short_message-8                	 7748691	       155 ns/op	     144 B/op	       3 allocs/op
BenchmarkLogg_Print/short_level-8                  	 8459841	       141 ns/op	     144 B/op	       3 allocs/op
BenchmarkLogg_Print/long_level-8                   	 8349727	       143 ns/op	     144 B/op	       3 allocs/op
BenchmarkLogg_Printf/long_message-8                	 3314443	       358 ns/op	    1154 B/op	       2 allocs/op
BenchmarkLogg_Printf/short_message-8               	 7906260	       153 ns/op	     128 B/op	       2 allocs/op
BenchmarkLogg_Printf/short_level-8                 	 8624889	       144 ns/op	     128 B/op	       2 allocs/op
BenchmarkLogg_Printf/long_level-8                  	 8253075	       145 ns/op	     128 B/op	       2 allocs/op
```


## Licence
[MIT License](https://github.com/pkgz/logg/blob/master/LICENSE)
