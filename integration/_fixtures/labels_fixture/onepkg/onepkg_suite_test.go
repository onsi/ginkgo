package onepkg

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOnepkg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Onepkg Suite")
}

var set1 = Label("dog", "cat", "cow")
