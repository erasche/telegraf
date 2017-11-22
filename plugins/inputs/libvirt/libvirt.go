package libvirt

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const sampleConfig = `
  ## specify a libvirt connection uri, see https://libvirt.org/uri.html
  uri = "qemu:///system"
`

type Libvirt struct {
	Uri   string
	virsh Virsh
}

type Virsh func(uri string, args ...string) (string, error)

func (l *Libvirt) SampleConfig() string {
	return sampleConfig
}

func (l *Libvirt) Description() string {
	return "Read domain infos from a libvirt deamon"
}

func (l *Libvirt) Gather(acc telegraf.Accumulator) error {
	domains, err := l.listDomains()

	if err != nil {
		return err
	}

	for _, domain := range domains {
		l.gatherDomain(acc, domain)
	}

	return nil
}

func (l *Libvirt) listDomains() ([]string, error) {
	out, err := l.virsh(l.Uri, "list")

	if err != nil {
		return []string{}, err
	}

	lines := strings.Split(out, "\n")

	domains := []string{}

	for _, line := range lines[2:] {
		if len(line) <= 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return []string{}, fmt.Errorf("failed to read domain list line: %s", line)
		}
		domains = append(domains, fields[1])
	}

	return domains, err
}

func runVirshCmd(uri string, cmd ...string) (string, error) {
	args := []string{"-c", uri}
	out, err := exec.Command("virsh", append(args, cmd...)...).Output()
	return string(out), err
}

func (l *Libvirt) gatherDomain(acc telegraf.Accumulator, domain string) error {

	out, err := l.virsh(l.Uri, "domstats", domain)

	if err != nil {
		return err
	}

	var fields = make(map[string]interface{})
	var state string

	for idx, line := range strings.Split(out, "\n") {
		if len(line) <= 0 {
			continue
		}
		if idx == 0 {
			continue
		}

		kv := strings.SplitN(line, "=", 2)

		if len(kv) != 2 {
			return fmt.Errorf("failed to read domain info for domain: %s, line: %q", domain, line)
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if (k == "state.state"){
			state = v
		}

		if (strings.HasSuffix(k, ".name") ) {
			continue
		} else {
			value, err := strconv.ParseUint(v, 0, 64)
			if err != nil {
				return err
			}
			if fields[k] == nil {
				fields[k] = 0
			}
			fields[k] = value

		}
	}

	tags := map[string]string{
		"domain": domain,
		"state":  state,
	}

	acc.AddFields("libvirt", fields, tags)

	return nil
}

func init() {
	inputs.Add("libvirt", func() telegraf.Input {
		return &Libvirt{
			virsh: runVirshCmd,
			Uri:   "qemu:///system",
		}
	})
}
