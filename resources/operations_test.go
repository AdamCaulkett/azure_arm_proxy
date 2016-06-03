package resources

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/ghttp"
	"github.com/rightscale/azure_arm_proxy/config"
)

const (
	listOneOperationResponse = `{"status":"succeeded","href":"locations/westus/operations/896da082-4e65-4d00-a1bc-8d86591949fc"}`
)

var _ = Describe("operations", func() {

	var do *ghttp.Server
	var client *AzureClient
	var response *Response
	var err error

	BeforeEach(func() {
		do = ghttp.NewServer()
		config.BaseURL = do.URL()
		client = NewAzureClient()
	})

	AfterEach(func() {
		do.Close()
	})

	Describe("get operation", func() {
		BeforeEach(func() {
			do.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/subscriptions/"+subscriptionID+"/providers/Microsoft.Compute/locations/westus/operations/896da082-4e65-4d00-a1bc-8d86591949fc"),
					ghttp.RespondWith(http.StatusOK, listOneOperationResponse),
				),
			)
			response, err = client.Get("/locations/westus/services/microsoft.compute/operations/896da082-4e65-4d00-a1bc-8d86591949fc")
		})

		It("no error occured", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns 200 status code", func() {
			Ω(do.ReceivedRequests()).Should(HaveLen(1))
			Ω(response.Status).Should(Equal(200))
		})

		It("returns a resource specific content type in the header", func() {
			Ω(response.Headers["Content-Type"][0]).Should(Equal("vnd.rightscale.operation+json"))
		})

		It("retrieves an existing instance", func() {
			var instance map[string]interface{}
			err := json.Unmarshal([]byte(listOneOperationResponse), &instance)
			Expect(err).NotTo(HaveOccurred())
			expected, err := json.Marshal(instance)
			Expect(err).NotTo(HaveOccurred())
			Ω(response.Body).Should(MatchJSON(expected))
		})
	})

	Describe("get operation with non-existant id", func() {
		BeforeEach(func() {
			do.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/subscriptions/"+subscriptionID+"/providers/Microsoft.Compute/locations/westus/operations/khrvi1"),
					ghttp.RespondWith(http.StatusNotFound, recordNotFound),
				),
			)
		})

		It("returns 200 but inform about issue in the details", func() {
			response, err = client.Get("/locations/westus/services/microsoft.compute/operations/khrvi1")
			Expect(err).NotTo(HaveOccurred())
			Ω(do.ReceivedRequests()).Should(HaveLen(1))
			Ω(response.Status).Should(Equal(200))
			Ω(response.Body).Should(Equal("{\"status\":\"failed\",\"details\":\"Error has occurred while requesting async operation: Could not find operation with id 'khrvi1'\",\"href\":\"locations/westus/operations/khrvi1\"}\n"))
		})
	})
})
