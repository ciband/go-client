// Copyright 2018 ArkEcosystem. All rights reserved.
//
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package ark_client

import (
	"bytes"
	"context"
	"encoding/json"
	// "errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	// "reflect"
	"strconv"
	"strings"
	"sync"
	// "time"
	// "github.com/google/go-querystring/query"
)

const (
	defaultBaseURL = "https://dexplorer.ark.io:8443/api/"
	userAgent      = "go-client"
)

type Client struct {
	clientMu sync.Mutex
	client   *http.Client

	BaseURL *url.URL

	common Service

	One_Accounts     *One_AccountsService
	One_Blocks       *One_BlocksService
	One_Delegates    *One_DelegatesService
	One_Loader       *One_LoaderService
	One_Peers        *One_PeersService
	One_Signatures   *One_SignaturesService
	One_Transactions *One_TransactionsService

	Two_Blocks       *Two_BlocksService
	Two_Delegates    *Two_DelegatesService
	Two_Node         *Two_NodeService
	Two_Peers        *Two_PeersService
	Two_Transactions *Two_TransactionsService
	Two_Votes        *Two_VotesService
	Two_Wallets      *Two_WalletsService
}

type Service struct {
	client *Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)

	c := &Client{client: httpClient, BaseURL: baseURL}
	c.common.client = c

	c.One_Accounts = (*One_AccountsService)(&c.common)
	c.One_Blocks = (*One_BlocksService)(&c.common)
	c.One_Delegates = (*One_DelegatesService)(&c.common)
	c.One_Loader = (*One_LoaderService)(&c.common)
	c.One_Peers = (*One_PeersService)(&c.common)
	c.One_Signatures = (*One_SignaturesService)(&c.common)
	c.One_Transactions = (*One_TransactionsService)(&c.common)

	c.Two_Blocks = (*Two_BlocksService)(&c.common)
	c.Two_Delegates = (*Two_DelegatesService)(&c.common)
	c.Two_Node = (*Two_NodeService)(&c.common)
	c.Two_Peers = (*Two_PeersService)(&c.common)
	c.Two_Transactions = (*Two_TransactionsService)(&c.common)
	c.Two_Votes = (*Two_VotesService)(&c.common)
	c.Two_Wallets = (*Two_WalletsService)(&c.common)

	return c
}

func (c *Client) SendRequest(ctx context.Context, version int, method string, urlStr string, body interface{}, v interface{}) (*http.Response, error) {
	// Create a new HTTP request
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("BaseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-Version", strconv.Itoa(version))

	// Execute the previously created HTTP request
	resp, err := c.client.Do(req)

	if err != nil {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if e, ok := err.(*url.Error); ok {
			if url, err := url.Parse(e.URL); err == nil {
				e.URL = url.String()
				return nil, e
			}
		}

		return nil, err
	}

	// Map the JSON response to a struct
	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, resp.Body)
		} else {
			decErr := json.NewDecoder(resp.Body).Decode(v)

			if decErr == io.EOF {
				decErr = nil
			}

			if decErr != nil {
				err = decErr
			}
		}
	}

	defer resp.Body.Close()

	return resp, nil
}
