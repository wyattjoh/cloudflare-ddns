package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
)

// GetRecord fetches the record from the Cloudflare api.
func GetRecord(api *cloudflare.API, domainName string) (*cloudflare.DNSRecord, error) {
	// Split the domain name by periods.
	splitDomainName := strings.Split(domainName, ".")

	// The domain name must be at least 2 elements, a name and a tld.
	if len(splitDomainName) < 2 {
		return nil, errors.Errorf("%s did not contain a TLD", domainName)
	}

	// Extract the zone name from the domain name. This should be the last two
	// period delimitered strings.
	zoneName := strings.Join(splitDomainName[len(splitDomainName)-2:], ".")

	// Fetch the zone ID
	zoneID, err := api.ZoneIDByName(zoneName) // Assuming example.com exists in your Cloudflare account already
	if err != nil {
		return nil, err
	}

	// Print zone details
	dnsRecords, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{
		Name: domainName,
	})
	if err != nil {
		return nil, err
	}

	if len(dnsRecords) != 1 {
		return nil, errors.Errorf("Expected to find a single dns record, got %d", len(dnsRecords))
	}

	// Capture the record id that we need to update.
	return &dnsRecords[0], nil
}

// GetCurrentIP gets the current machine's external IP address from the
// https://ipify.org service.
func GetCurrentIP(ipEndpoint string) (string, error) {
	resp, err := http.Get(ipEndpoint)
	if err != nil {
		return "", errors.Wrap(err, "could not get the current IP from the provider")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read the output from the provider")
	}

	// Update the IP address.
	return string(data), nil
}

// UpdateDomain updates a given domain in a zone to match the current ip address
// of the machine.
func UpdateDomain(apiKey, apiEmail, domainNames, ipEndpoint string) error {
	// Create the new Cloudflare api client.
	api, err := cloudflare.New(apiKey, apiEmail)
	if err != nil {
		return errors.Wrap(err, "could not create the Cloudflare API client")
	}

	// Get our current IP address.
	newIP, err := GetCurrentIP(ipEndpoint)
	if err != nil {
		return errors.Wrap(err, "could not get the current IP address")
	}

	// Split the domain names by comma, and range over them.
	splitDomainNames := strings.Split(domainNames, ",")
	for _, domainName := range splitDomainNames {
		// Get the record in question.
		record, err := GetRecord(api, domainName)
		if err != nil {
			return errors.Wrap(err, "could not get the DNS record")
		}

		// Update the DNS record to include the new IP address.
		record.Content = newIP

		if err := api.UpdateDNSRecord(record.ZoneID, record.ID, *record); err != nil {
			return errors.Wrap(err, "could not update the DNS record")
		}

		// Log the update.
		fmt.Printf("Updated %s to point to %s\n", record.Name, record.Content)
	}

	return nil
}

func main() {
	// Extract the configuration from the environment.
	var APIKey, APIEmail, DomainNames, IPEndpoint string

	// Specify a default endpoint if no other one is provided.
	const defaultIPEndpoint = "https://api.ipify.org/"

	IPEndpoint = os.Getenv("CF_IP_ENDPOINT")

	// Default to the defaultIPEndpoint if no alternative was specified.
	if IPEndpoint == "" {
		IPEndpoint = defaultIPEndpoint
	}

	flags := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Define the arguments needed.
	flags.StringVar(&APIKey, "key", os.Getenv("CF_API_KEY"), "specify the Global (not CA) Cloudflare API Key generated on the \"My Account\" page.")
	flags.StringVar(&APIEmail, "email", os.Getenv("CF_API_EMAIL"), "Email address associated with your Cloudflare account.")
	flags.StringVar(&DomainNames, "domain", os.Getenv("CF_DOMAIN"), "Comma separated domain names that should be updated. (i.e. mypage.example.com OR example.com)")
	flags.StringVar(&IPEndpoint, "ipendpoint", IPEndpoint, "Alternative ip address service endpoint.")

	// Parse the flags in.
	if err := flags.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			// Error nicely if it was just asking for help.
			os.Exit(0)
		}

		// Exit not nicely otherwise.
		os.Exit(2)
	}

	if err := UpdateDomain(APIKey, APIEmail, DomainNames, IPEndpoint); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
