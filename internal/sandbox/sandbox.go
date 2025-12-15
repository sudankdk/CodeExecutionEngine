package sandbox

import "time"

type Config struct {
	ExecCmd        []string
	Stdin          string
	Memory         int64 // bytes
	CPU            int64 // NanoCPUs
	Binds          []string
	ReadonlyRootfs bool
	Timeout        time.Duration
}

func NewConfig(codeDir, codeFilename, stdinFilename, lang string) Config {
	var cmd []string
	readonly := true
	timeout := 20 * time.Second

	switch lang {
	case "python":
		cmd = []string{"python3", codeFilename}
	case "go":
		cmd = []string{"go", "run", codeFilename}
		readonly = false
		timeout = 40 * time.Second
	case "node":
		cmd = []string{"node", codeFilename}
	}

	return Config{
		ExecCmd:        cmd,
		Memory:         512 * 1024 * 1024,
		CPU:            1_000_000_000,
		Binds:          []string{codeDir + ":/run/code"},
		ReadonlyRootfs: readonly,
		Timeout:        timeout,
	}
}

// NewConfig builds a sandbox config for a language run.
// codeDir: host dir containing main<ext> and stdin.txt
// codeFilename: relative path inside container (e.g. /run/main.py)
// func NewConfig(codeDir, codeFilename, stdinFilename, lang string) Config {
// 	var cmd []string
// 	readonly := true
// 	timeout := 20 * time.Second
// 	switch lang {
// 	case "python":
// 		cmd = []string{codeFilename, stdinFilename}
// 	case "go":
// 		cmd = []string{codeFilename, stdinFilename}
// 		readonly = false
// 		timeout = 40 * time.Second
// 	case "node":
// 		cmd = []string{codeFilename, stdinFilename}
// 	}

// 	return Config{
// 		Cmd:            cmd,
// 		Memory:         512 * 1024 * 1024,
// 		CPU:            1000000000,
// 		Binds:          []string{codeDir + ":/run/code"},
// 		ReadonlyRootfs: readonly,
// 		Timeout:        timeout,
// 	}
// }
