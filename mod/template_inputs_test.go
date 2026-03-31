package dockergatusgenerator

import (
	"os"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func assertContainer(t *testing.T, ti TemplateInput, pos int, id string) bool {
	assert.LessOrEqual(t, pos+1, len(ti.Containers), "There is no element %d", pos)
	assert.Equal(t, id, ti.Containers[pos].ID)
	return true
}

func makeTestData() TemplateInputs {
	ti := TemplateInputs{}
	ti.Append("test", Container{
		ID:   "1",
		Name: "best",
	})
	ti.Append("test", Container{
		ID:   "2",
		Name: "aubrey",
	})
	ti.Append("test", Container{
		ID:   "3",
		Name: "of",
	})
	ti.Append("test", Container{
		ID:   "4",
		Name: "the",
	})
	ti.Append("test", Container{
		ID:   "5",
		Name: "world",
	})
	ti.Append("pruebita", Container{
		ID:   "a",
		Name: "helo",
	})
	ti.Append("pruebita", Container{
		ID:   "b",
		Name: "world",
	})
	return ti
}

func TestSortsInputsByName(t *testing.T) {
	ti := makeTestData()

	ti.Finish(TestConfig{})

	c := ti["test"]
	assertContainer(t, c, 0, "2")
	assertContainer(t, c, 1, "1")
	assertContainer(t, c, 2, "3")
	assertContainer(t, c, 3, "4")
	assertContainer(t, c, 4, "5")

	c = ti["pruebita"]
	assertContainer(t, c, 0, "a")
	assertContainer(t, c, 1, "b")
}

func TestFillsWithHostnameAndIpFromConfig(t *testing.T) {
	ti := makeTestData()

	ti.Finish(TestConfig{ENV_GATUS_HOSTNAME: "computer.local", ENV_GATUS_IP: "192.168.1.1"})

	c := ti["test"]
	assert.Equal(t, "computer.local", c.Hostname)
	assert.Equal(t, "192.168.1.1", c.Ip)
}

func TestDefaultsToMachineHostnameIfConfigIsEmpty(t *testing.T) {
	ti := makeTestData()

	ti.Finish(TestConfig{ENV_GATUS_IP: "192.168.1.1"})

	c := ti["test"]
	if h, ok := os.Hostname(); ok == nil {
		assert.Equal(t, h, c.Hostname)
	}
	assert.Equal(t, "192.168.1.1", c.Ip)
}
