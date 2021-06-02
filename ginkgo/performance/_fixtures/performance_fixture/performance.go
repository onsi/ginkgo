package performance

import (
	"fmt"

	"github.com/gorilla/mux"
	"github.com/tdewolff/minify/v2"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
)

func main() {
	mux.NewRouter()
	fmt.Println(bcrypt.MinCost)
	fmt.Println(yaml.Decoder{})
	fmt.Println(minify.MinInt)
}
