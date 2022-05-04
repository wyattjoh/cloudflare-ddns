package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// GetRecord fetches the record from the Cloudflare api.
func GetRecord(ctx context.Context, api *cloudflare.API, domainName string) (*cloudflare.DNSRecord, error) {
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
		return nil, errors.Wrap(err, "could not find zone by name")
	}

	logrus.WithField("zoneID", zoneID).Debug("got zone id")

	// Print zone details
	dnsRecords, err := api.DNSRecords(ctx, zoneID, cloudflare.DNSRecord{
		Name: domainName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not locate dns record for zone")
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
	req, err := http.NewRequest("GET", ipEndpoint, nil)
	if err != nil {
		return "", errors.Wrap(err, "could not create the request to the IP provider")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "could not get the current IP from the provider")
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", errors.Wrap(err, "could not read the output from the provider")
	}

	// Update the IP address.
	return string(data), nil
}

// UpdateDomain updates a given domain in a zone to match the current ip address
// of the machine.
func UpdateDomain(ctx context.Context, api *cloudflare.API, domainNames, ipEndpoint string) error {
	// // Create the new Cloudflare api client.
	// cloudflare.NewWithAPIToken()
	// api, err := cloudflare.New(apiKey, apiEmail)
	// if err != nil {
	// 	return errors.Wrap(err, "could not create the Cloudflare API client")
	// }

	// Get our current IP address.
	newIP, err := GetCurrentIP(ipEndpoint)
	if err != nil {
		return errors.Wrap(err, "could not get the current IP address")
	}

	logrus.WithField("ip", newIP).Debug("got current IP address")

	// Split the domain names by comma, and range over them.
	splitDomainNames := strings.Split(domainNames, ",")
	for _, domainName := range splitDomainNames {
		// Get the record in question.
		record, err := GetRecord(ctx, api, domainName)
		if err != nil {
			return errors.Wrap(err, "could not get the DNS record")
		}

		// Update the DNS record to include the new IP address.
		record.Content = newIP

		if err := api.UpdateDNSRecord(ctx, record.ZoneID, record.ID, *record); err != nil {
			return errors.Wrap(err, "could not update the DNS record")
		}

		// Log the update.
		logrus.WithFields(logrus.Fields{
			"name":    record.Name,
			"content": record.Content,
		}).Info("updated record")
	}

	return nil
}
