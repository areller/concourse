package tsa_test

import (
	"context"

	"github.com/concourse/concourse/tsa"
	"golang.org/x/oauth2"

	"code.cloudfoundry.org/lager/v3/lagerctx"
	"code.cloudfoundry.org/lager/v3/lagertest"
	"github.com/concourse/concourse/atc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/tedsuo/rata"
)

var _ = Describe("Lander", func() {
	var (
		lander *tsa.Lander

		ctx     context.Context
		worker  atc.Worker
		fakeATC *ghttp.Server
	)

	BeforeEach(func() {
		ctx = lagerctx.NewContext(context.Background(), lagertest.NewTestLogger("test"))
		worker = atc.Worker{
			Name: "some-worker",
		}
		fakeATC = ghttp.NewServer()

		atcEndpoint := rata.NewRequestGenerator(fakeATC.URL(), atc.Routes)

		token := &oauth2.Token{TokenType: "Bearer", AccessToken: "yo"}
		httpClient := oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(token))

		lander = &tsa.Lander{
			ATCEndpoint: atcEndpoint,
			HTTPClient:  httpClient,
		}
	})

	AfterEach(func() {
		fakeATC.Close()
	})

	Context("when the worker request is for a team-owned worker", func() {
		BeforeEach(func() {
			worker.Team = "some-team"
		})

		It("tells the ATC to land the worker", func() {
			fakeATC.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/workers/some-worker/land"),
				ghttp.VerifyHeaderKV("Authorization", "Bearer yo"),
				ghttp.RespondWith(200, nil, nil),
			))

			err := lander.Land(ctx, worker)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
		})
	})

	It("tells the ATC to land the worker", func() {
		fakeATC.AppendHandlers(ghttp.CombineHandlers(
			ghttp.VerifyRequest("PUT", "/api/v1/workers/some-worker/land"),
			ghttp.VerifyHeaderKV("Authorization", "Bearer yo"),
			ghttp.RespondWith(200, nil, nil),
		))

		err := lander.Land(ctx, worker)
		Expect(err).NotTo(HaveOccurred())

		Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
	})

	Context("when the ATC responds with a 403", func() {
		BeforeEach(func() {
			fakeATC.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/workers/some-worker/land"),
				ghttp.RespondWith(403, nil, nil),
			))
		})

		It("errors", func() {
			err := lander.Land(ctx, worker)
			Expect(err).To(HaveOccurred())

			Expect(err).To(MatchError(ContainSubstring("403")))
			Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Context("when the ATC responds with a 404", func() {
		BeforeEach(func() {
			fakeATC.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/workers/some-worker/land"),
				ghttp.RespondWith(404, nil, nil),
			))
		})

		It("errors", func() {
			err := lander.Land(ctx, worker)
			Expect(err).To(HaveOccurred())

			Expect(err).To(MatchError(ContainSubstring("404")))
			Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
		})
	})

	Context("when the ATC fails to land the worker", func() {
		BeforeEach(func() {
			fakeATC.AppendHandlers(ghttp.CombineHandlers(
				ghttp.VerifyRequest("PUT", "/api/v1/workers/some-worker/land"),
				ghttp.RespondWith(500, nil, nil),
			))
		})

		It("errors", func() {
			err := lander.Land(ctx, worker)
			Expect(err).To(HaveOccurred())

			Expect(err).To(MatchError(ContainSubstring("500")))
			Expect(fakeATC.ReceivedRequests()).To(HaveLen(1))
		})
	})
})
