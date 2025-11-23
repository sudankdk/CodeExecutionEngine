package executer

type Config struct{
	Cmd []string
	Memory int64
	Cpu int64
	Binds []string
}

func NewConfig(code,stdin string) Config {
	return  Config{
		Cmd: []string{"/entrypoint.sh"},
		Memory: 256 * 1024 * 1024,
		Cpu: 500_000_000,  //0.5 cpu hunxa
		Binds: []string{
            "/tmp/code:/run/code:ro",
            "/tmp/stdin:/run/stdin:ro",
		},
	}
}