package sandbox

type Config struct {
	Cmd           []string
	Memory        int64  // bytes
	CPU           int64  // NanoCPUs
	Binds         []string
	ReadonlyRootfs bool
}

// NewConfig builds a sandbox config for a language run.
// codeDir: host dir containing main<ext> and stdin.txt
// codeFilename: relative path inside container (e.g. /run/main.py)
func NewConfig(codeDir, codeFilename, stdinFilename, lang string) Config {
    var cmd []string
    readonly:= true
    switch lang {
    case "python":
        cmd = []string{ codeFilename, stdinFilename}
    case "go":
        cmd = []string{codeFilename, stdinFilename}
        readonly=false
    case "node":
        cmd = []string{codeFilename ,stdinFilename}
    }

    return Config{
        Cmd:            cmd,
        Memory:         512 * 1024 * 1024,
        CPU:            500_000_000,
        Binds:          []string{codeDir + ":/run/code:ro"},
        ReadonlyRootfs: readonly,
    }
}

