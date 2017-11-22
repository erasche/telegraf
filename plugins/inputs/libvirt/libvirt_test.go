package libvirt

import (
	"fmt"
	"reflect"
	"testing"

	"errors"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

const mockListStr = ` Id    Name                           State
----------------------------------------------------
 1     vm1                           running
 2     vm2                           running

`

const mockListNoVmsStr = ` Id    Name                           State
----------------------------------------------------

`

const mockListNotInstalled = `The program 'virsh' is currently not installed. You can install it by typing:
sudo apt install libvirt-bin
`

const mockDomInfo1 = `Domain: 'os_e4aa2166-1fc2-40da-bc7b-bdee9d4b46e0'
  state.state=1
  state.reason=1
  cpu.time=27922986816134
  cpu.user=12435590000000
  cpu.system=9190820000000
  balloon.current=8388608
  balloon.maximum=8388608
  vcpu.current=8
  vcpu.maximum=8
  vcpu.0.state=1
  vcpu.0.time=503450000000
  vcpu.0.wait=0
  vcpu.1.state=1
  vcpu.1.time=340360000000
  vcpu.1.wait=0
  vcpu.2.state=1
  vcpu.2.time=112240000000
  vcpu.2.wait=0
  vcpu.3.state=1
  vcpu.3.time=151240000000
  vcpu.3.wait=0
  vcpu.4.state=1
  vcpu.4.time=143950000000
  vcpu.4.wait=0
  vcpu.5.state=1
  vcpu.5.time=1777240000000
  vcpu.5.wait=0
  vcpu.6.state=1
  vcpu.6.time=315410000000
  vcpu.6.wait=0
  vcpu.7.state=1
  vcpu.7.time=117930000000
  vcpu.7.wait=0
  net.count=1
  net.0.name=tap59ebedbf-aa
  net.0.rx.bytes=578192543
  net.0.rx.pkts=6767928
  net.0.rx.errs=0
  net.0.rx.drop=0
  net.0.tx.bytes=9055661
  net.0.tx.pkts=52624
  net.0.tx.errs=0
  net.0.tx.drop=0
  block.count=1
  block.0.name=vda
  block.0.rd.reqs=20285
  block.0.rd.bytes=333184000
  block.0.rd.times=21651023700
  block.0.wr.reqs=113478
  block.0.wr.bytes=3330606080
  block.0.wr.times=659537320620
  block.0.fl.reqs=70008
  block.0.fl.times=5021028670
  block.0.allocation=19328466944
  block.0.capacity=21474836480
  block.0.physical=21474836480

`

const mockDomInfo2 = `Domain: 'os_d12f2b17-a853-45d6-8ab8-8a11678ebf13'
  state.state=1
  state.reason=1
  cpu.time=456175352624347
  cpu.user=127085640000000
  cpu.system=109092970000000
  balloon.current=4194304
  balloon.maximum=4194304
  vcpu.current=2
  vcpu.maximum=2
  vcpu.0.state=1
  vcpu.0.time=88070380000000
  vcpu.0.wait=0
  vcpu.1.state=1
  vcpu.1.time=108807740000000
  vcpu.1.wait=0
  net.count=1
  net.0.name=tap44b015bf-2a
  net.0.rx.bytes=59877546125
  net.0.rx.pkts=142259305
  net.0.rx.errs=0
  net.0.rx.drop=1034
  net.0.tx.bytes=2412227400358
  net.0.tx.pkts=74847914
  net.0.tx.errs=0
  net.0.tx.drop=0
  block.count=2
  block.0.name=vda
  block.0.rd.reqs=48027
  block.0.rd.bytes=1423333376
  block.0.rd.times=64180669466
  block.0.wr.reqs=2186923
  block.0.wr.bytes=32228062208
  block.0.wr.times=11084560286151
  block.0.fl.reqs=923050
  block.0.fl.times=154339835035
  block.0.allocation=8810217472
  block.0.capacity=10737418240
  block.0.physical=10737418240
  block.1.name=vdb
  block.1.rd.reqs=120292
  block.1.rd.bytes=14896106496
  block.1.rd.times=166008983815
  block.1.wr.reqs=47051
  block.1.wr.bytes=6167621632
  block.1.wr.times=504061692264
  block.1.fl.reqs=31332
  block.1.fl.times=3963274627
  block.1.allocation=36641816576
  block.1.capacity=53687091200
  block.1.physical=53687091200

`

func mockVirshNormal(_ string, cmd ...string) (string, error) {
	if reflect.DeepEqual([]string{"list"}, cmd) {
		return mockListStr, nil
	} else if reflect.DeepEqual([]string{"dominfo", "vm1"}, cmd) {
		return mockDomInfo1, nil
	} else if reflect.DeepEqual([]string{"dominfo", "vm2"}, cmd) {
		return mockDomInfo2, nil
	}

	return "", fmt.Errorf("unknown cmd: %q", cmd)
}

func mockVirshNoVms(_ string, cmd ...string) (string, error) {
	if reflect.DeepEqual([]string{"list"}, cmd) {
		return mockListNoVmsStr, nil
	}

	return "", fmt.Errorf("unexpected cmd: %q", cmd)
}

func mockVirshNotInstalled(_ string, cmd ...string) (string, error) {
	if reflect.DeepEqual([]string{"list"}, cmd) {
		return mockListNotInstalled, errors.New("exec: executable file not found in $PATH")
	}

	return "", fmt.Errorf("unexpected cmd: %q", cmd)
}

func TestLibvirtNormal(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default", virsh: mockVirshNormal}

	err := lv.Gather(&acc)
	require.NoError(t, err)

	vm1_tags := map[string]string{
		"domain": "vm1",
		"state":  "running",
	}
	vm1_fields := map[string]interface{}{
		"cpu_time":    1489945053.2,
		"max_memory":  uint64(8388608),
		"used_memory": uint64(2097152),
		"n_vcpu":      uint64(2),
	}

	vm2_tags := map[string]string{
		"domain": "vm2",
		"state":  "running",
	}
	vm2_fields := map[string]interface{}{
		"cpu_time":    11234.6,
		"max_memory":  uint64(4194304),
		"used_memory": uint64(4194304),
		"n_vcpu":      uint64(8),
	}

	acc.AssertContainsTaggedFields(t, "libvirt", vm1_fields, vm1_tags)
	acc.AssertContainsTaggedFields(t, "libvirt", vm2_fields, vm2_tags)
}

func TestLibvirtNoVms(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default", virsh: mockVirshNoVms}

	err := lv.Gather(&acc)
	require.NoError(t, err)

	acc.AssertDoesNotContainMeasurement(t, "libvirt")
}

func TestLibvirtNotInstalled(t *testing.T) {
	var acc testutil.Accumulator

	lv := &Libvirt{Uri: "test:///default", virsh: mockVirshNotInstalled}

	err := lv.Gather(&acc)
	require.EqualError(t, err, "exec: executable file not found in $PATH", "expect not installed to fail")

	acc.AssertDoesNotContainMeasurement(t, "libvirt")
}
