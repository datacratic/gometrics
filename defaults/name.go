package defaults

import (
	"os"
	"os/exec"
	"strings"
	"sync"
)

var fqdn = struct {
	once sync.Once
	name string
}{}

// Name returns the default service name.
// It's built from the process name and the (reverse) fully qualified domain name.
func Name() string {
	fqdn.once.Do(func() {
		// start with the process name from the command line
		name := os.Args[0]
		if i := strings.LastIndex(name, "/"); i >= 0 {
			name = name[i+1:]
		}

		// include the fqdn if possible
		if hostname, err := exec.Command("hostname", "--fqdn").Output(); err == nil {
			h := string(hostname[:len(hostname)-1])
			p := strings.Split(h, ".")

			// use reverse notation
			for i, j := 0, len(p)-1; i < j; i, j = i+1, j-1 {
				p[i], p[j] = p[j], p[i]
			}

			name = strings.Join(p, ".") + "." + name
		}

		fqdn.name = name
	})

	return fqdn.name
}
