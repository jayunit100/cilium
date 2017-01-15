//
// Copyright 2016-2017 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package client

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	clientapi "github.com/cilium/cilium/api/v1/client"
	common "github.com/cilium/cilium/common"
	runtime_client "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

type Client struct {
	clientapi.Cilium
}

func configureTransport(tr *http.Transport, proto, addr string) *http.Transport {
	if tr == nil {
		tr = &http.Transport{}
	}

	if proto == "unix" {
		// No need for compression in local communications.
		tr.DisableCompression = true
		tr.Dial = func(_, _ string) (net.Conn, error) {
			return net.Dial(proto, addr)
		}
	} else {
		tr.Proxy = http.ProxyFromEnvironment
		tr.Dial = (&net.Dialer{}).Dial
	}

	return tr
}

func NewClient(host string, transport *http.Transport) (*Client, error) {
	if host == "" {
		host = "unix://" + common.CiliumSock
	}

	tmp := strings.SplitN(host, "://", 2)
	if len(tmp) != 2 {
		return nil, fmt.Errorf("invalid host format '%s'", host)
	}

	proto, addr := tmp[0], tmp[1]

	switch proto {
	case "tcp":
		if _, err := url.Parse("tcp://" + addr); err != nil {
			return nil, err
		}
		addr = "http://" + addr
	case "http":
		addr = "http://" + addr
	}

	transport = configureTransport(transport, proto, addr)
	httpClient := &http.Client{Transport: transport}
	clientTrans := runtime_client.NewWithClient(host, "/v1", []string{"http"}, httpClient)
	return &Client{*clientapi.New(clientTrans, strfmt.Default)}, nil
}
