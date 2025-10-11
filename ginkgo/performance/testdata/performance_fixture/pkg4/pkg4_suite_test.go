package pkg4_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tdewolff/minify/v2"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
)

func TestPkg4(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg4 Suite")

	mux.NewRouter()
	fmt.Println(bcrypt.MinCost)
	fmt.Println(yaml.Decoder{})
	fmt.Println(minify.MinInt)
}

var _ = Describe("Pkg4", func() {
	for i := 0; i < 10; i++ {
		It(fmt.Sprintf("sleeps %d", i), func() {
			time.Sleep(time.Millisecond * 10)
		})
	}
})
