package pkg1_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tdewolff/minify/v2"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
)

func TestPkg1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pkg1 Suite")

	mux.NewRouter()
	fmt.Println(bcrypt.MinCost)
	fmt.Println(yaml.Decoder{})
	fmt.Println(minify.MinInt)
}

var _ = Describe("Pkg1", func() {
	for i := 0; i < 10; i++ {
		It(fmt.Sprintf("sleeps %d", i), func() {
			time.Sleep(time.Millisecond * 10)
		})
	}
})
