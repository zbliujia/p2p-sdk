package mypkg

import (
	"bufio"
	"context"
	"fmt"
	ma "github.com/multiformats/go-multiaddr"
	manet "github.com/multiformats/go-multiaddr/net"
	"github.com/zbliujia/go-libp2p"
	"github.com/zbliujia/go-libp2p/core/host"
	"github.com/zbliujia/go-libp2p/core/network"
	"github.com/zbliujia/go-libp2p/core/peer"
	"github.com/zbliujia/go-libp2p/core/peerstore"
	"io"
	"log"
	"net/http"
	"strings"
)

func makeRandomHost(address string, port int) host.Host {
	h, err := libp2p.New(libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/%s/tcp/%d", address, port)))
	if err != nil {
		log.Fatalln(err)
	}
	return h
}

func addAddrToPeerstore(h host.Host, addr string) peer.ID {
	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		log.Fatalln(err)
	}
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.Decode(pid)
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peerid))
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	// We have a peer ID and a targetAddr, so we add
	// it to the peerstore so LibP2P knows how to contact it
	h.Peerstore().AddAddr(peerid, targetAddr, peerstore.PermanentAddrTTL)
	return peerid
}

type ProxyService struct {
	host      host.Host
	dest      peer.ID
	proxyAddr ma.Multiaddr
}

const Protocol = "/proxy-example/0.0.1"

func streamHandler(stream network.Stream) {
	// Remember to close the stream when we are done.
	defer stream.Close()

	// Create a new buffered reader, as ReadRequest needs one.
	// The buffered reader reads from our stream, on which we
	// have sent the HTTP request (see ServeHTTP())
	buf := bufio.NewReader(stream)
	// Read the HTTP request from the buffer
	req, err := http.ReadRequest(buf)
	if err != nil {
		stream.Reset()
		log.Println(err)
		return
	}
	defer req.Body.Close()

	// We need to reset these fields in the request
	// URL as they are not maintained.
	req.URL.Scheme = "http"
	hp := strings.Split(req.Host, ":")
	if len(hp) > 1 && hp[1] == "443" {
		req.URL.Scheme = "https"
	} else {
		req.URL.Scheme = "http"
	}
	req.URL.Host = req.Host

	outreq := new(http.Request)
	*outreq = *req

	// We now make the request
	fmt.Printf("Making request to %s\n", req.URL)
	resp, err := http.DefaultTransport.RoundTrip(outreq)
	if err != nil {
		stream.Reset()
		log.Println(err)
		return
	}

	// resp.Write writes whatever response we obtained for our
	// request back to the stream.
	resp.Write(stream)
}

func (p *ProxyService) Serve() {
	_, serveArgs, _ := manet.DialArgs(p.proxyAddr)
	fmt.Println("proxy listening on ", serveArgs)
	if p.dest != "" {
		http.ListenAndServe(serveArgs, p)
	}
}

func (p *ProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("proxying request for %s to peer %s\n", r.URL, p.dest)
	// We need to send the request to the remote libp2p peer, so
	// we open a stream to it
	stream, err := p.host.NewStream(context.Background(), p.dest, Protocol)
	// If an error happens, we write an error for response.
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stream.Close()

	// r.Write() writes the HTTP request to the stream.
	err = r.Write(stream)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Now we read the response that was sent from the dest
	// peer
	buf := bufio.NewReader(stream)
	resp, err := http.ReadResponse(buf, r)
	if err != nil {
		stream.Reset()
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Copy any headers
	for k, v := range resp.Header {
		for _, s := range v {
			w.Header().Add(k, s)
		}
	}

	// Write response status and headers
	w.WriteHeader(resp.StatusCode)

	// Finally copy the body
	io.Copy(w, resp.Body)
	resp.Body.Close()
}

func newProxyService(h host.Host, proxyAddr ma.Multiaddr, dest peer.ID) *ProxyService {
	// We let our host know that it needs to handle streams tagged with the
	// protocol id that we have defined, and then handle them to
	// our own streamHandling function.
	h.SetStreamHandler(Protocol, streamHandler)

	fmt.Println("Proxy server is ready")
	fmt.Println("libp2p-peer addresses:")
	for _, a := range h.Addrs() {
		fmt.Printf("%s/ipfs/%s\n", a, h.ID())
	}

	return &ProxyService{
		host:      h,
		dest:      dest,
		proxyAddr: proxyAddr,
	}
}
