package leafnodes_test

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/internal/leafnodes"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("CompoundServer", func() {
	var server *CompoundServer
	var err error

	BeforeEach(func() {
		server, err = NewCompoundServer(3)
		Ω(err).ShouldNot(HaveOccurred())
		server.Start()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("GETting and POSTing BeforeSuiteState", func() {
		getBeforeSuite := func() RemoteState {
			resp, err := http.Get("http://" + server.Address() + "/BeforeSuiteState")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp.StatusCode).Should(Equal(http.StatusOK))

			r := RemoteState{}
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&r)
			Ω(err).ShouldNot(HaveOccurred())

			return r
		}

		postBeforeSuite := func(r RemoteState) {
			resp, err := http.Post("http://"+server.Address()+"/BeforeSuiteState", "application/json", bytes.NewReader(r.ToJSON()))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp.StatusCode).Should(Equal(http.StatusOK))
		}

		Context("when the first node's Alive has not been registered yet", func() {
			It("should return pending", func() {
				state := getBeforeSuite()
				Ω(state).Should(Equal(RemoteState{nil, RemoteStateStatePending}))

				state = getBeforeSuite()
				Ω(state).Should(Equal(RemoteState{nil, RemoteStateStatePending}))
			})
		})

		Context("when the first node is Alive but has not responded yet", func() {
			BeforeEach(func() {
				server.RegisterAlive(1, func() bool {
					return true
				})
			})

			It("should return pending", func() {
				state := getBeforeSuite()
				Ω(state).Should(Equal(RemoteState{nil, RemoteStateStatePending}))

				state = getBeforeSuite()
				Ω(state).Should(Equal(RemoteState{nil, RemoteStateStatePending}))
			})
		})

		Context("when the first node has responded", func() {
			var state RemoteState
			BeforeEach(func() {
				server.RegisterAlive(1, func() bool {
					return false
				})

				state = RemoteState{
					Data:  []byte("my data"),
					State: RemoteStateStatePassed,
				}
				postBeforeSuite(state)
			})

			It("should return the passed in state", func() {
				returnedState := getBeforeSuite()
				Ω(returnedState).Should(Equal(state))
			})
		})

		Context("when the first node is no longer Alive and has not responded yet", func() {
			BeforeEach(func() {
				server.RegisterAlive(1, func() bool {
					return false
				})
			})

			It("should return disappeared", func() {
				state := getBeforeSuite()
				Ω(state).Should(Equal(RemoteState{nil, RemoteStateStateDisappeared}))

				state = getBeforeSuite()
				Ω(state).Should(Equal(RemoteState{nil, RemoteStateStateDisappeared}))
			})
		})
	})

	Describe("GETting AfterSuiteCanRun", func() {
		getAfterSuiteCanRun := func() bool {
			resp, err := http.Get("http://" + server.Address() + "/AfterSuiteCanRun")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(resp.StatusCode).Should(Equal(http.StatusOK))

			a := AfterSuiteCanRun{}
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&a)
			Ω(err).ShouldNot(HaveOccurred())

			return a.CanRun
		}

		Context("when there are unregistered nodes", func() {
			BeforeEach(func() {
				server.RegisterAlive(2, func() bool {
					return false
				})
			})

			It("should return false", func() {
				Ω(getAfterSuiteCanRun()).Should(BeFalse())
			})
		})

		Context("when all none-node-1 nodes are still running", func() {
			BeforeEach(func() {
				server.RegisterAlive(2, func() bool {
					return true
				})

				server.RegisterAlive(3, func() bool {
					return false
				})
			})

			It("should return false", func() {
				Ω(getAfterSuiteCanRun()).Should(BeFalse())
			})
		})

		Context("when all none-1 nodes are done", func() {
			BeforeEach(func() {
				server.RegisterAlive(2, func() bool {
					return false
				})

				server.RegisterAlive(3, func() bool {
					return false
				})
			})

			It("should return true", func() {
				Ω(getAfterSuiteCanRun()).Should(BeTrue())
			})

		})
	})
})
