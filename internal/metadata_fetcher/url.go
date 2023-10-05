package metadata_fetcher

import (
	"fmt"
	"gitlab.com/distributed_lab/logan/v3/errors"
	"net/url"
	"strings"
)

func wrapUrl(url string, nftId string) string {
	if !strings.HasPrefix(url, "https://api.opensea.io") {
		return url
	}

	// URL with 0x don't work
	return strings.Replace(url, "0x{id}", nftId, 1)
}

func interpolateUrl(url string, nftId string) string {
	return strings.Replace(url, "{id}", nftId, 1)
}

var publicIpfsGateways = []string{
	"https://ipfs.io/ipfs/",
	"https://gateway.pinata.cloud/ipfs/",
	"https://cloudflare-ipfs.com/ipfs/",
	"https://gateway.ipfs.io/ipfs/",
	"https://dweb.link/ipfs/",
	"https://hardbin.com/ipfs/",
	"https://ipfs.fleek.co/ipfs/",
	"https://jorropo.net/ipfs/",
	"https://ipfs.eth.aragon.network/ipfs/",
	"https://cloudflare-ipfs.com/ipfs/",
	"https://storry.tv/ipfs/",
	"https://ipfs.telos.miami/ipfs/",
	"https://via0.com/ipfs/",
	"https://ipfs.infura.io/ipfs/",
	"https://infura-ipfs.io/ipfs/",
	"https://ipfs.mihir.ch/ipfs/",
	"https://nftstorage.link/ipfs/",
	"https://cf-ipfs.com/ipfs/",
	"https://gateway.pinata.cloud/ipfs/",
	"https://ipfs.azurewebsites.net/ipfs/",
	"https://permaweb.eu.org/ipfs/",
}

var httpPrefix = "http"
var httpsPrefix = "https"

// StripIPFSGateway - strips public IPFS gateways URLs, and returns URL in IPFS protocol format
func stripIPFSGateway(uri string) string {
	if !strings.HasPrefix(uri, httpPrefix) {
		return uri
	}

	sourceUrl := uri

	if !strings.HasPrefix(uri, httpsPrefix) {
		sourceUrl = fmt.Sprintf("%s://", httpsPrefix) + strings.ReplaceAll(uri, fmt.Sprintf("%s://", httpPrefix), "")
	}

	for _, ipfsGatewayToReplace := range publicIpfsGateways {
		if strings.HasPrefix(sourceUrl, ipfsGatewayToReplace) {
			return "ipfs://" + strings.ReplaceAll(sourceUrl, ipfsGatewayToReplace, "")
		}
	}

	return uri
}

func parseUri(uri, nftId string) (*url.URL, error) {
	uri = wrapUrl(uri, nftId)
	uri = interpolateUrl(uri, nftId)

	parsedURL, err := url.Parse(stripIPFSGateway(uri))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse uri")
	}

	return parsedURL, nil
}
