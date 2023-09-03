package requester

import (
	"fmt"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/hashicorp/go-multierror"
	"sync"
)

func Initialize() {
	// load package to launch init function
}

func FetchRequests(requesters types.Requesters) ([]*types.DomainRequest, error) {
	merr := &multierror.Error{}
	mtx := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	domainsRequests := []*types.DomainRequest{}
	for _, requester := range requesters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			domainsRequest, err := requester.Fetch()
			if err != nil {
				merr = multierror.Append(merr, fmt.Errorf(
					"http request (%s): requester (%s) fetching domains request failed: %v",
					appHttp.GetApiPrefix(appHttp.AgentApiRequests),
					requester.ID(),
					err,
				))
			}
			mtx.Lock()
			defer mtx.Unlock()
			domainsRequests = append(domainsRequests, domainsRequest...)
		}()

	}
	wg.Wait()
	return domainsRequests, merr.ErrorOrNil()
}
