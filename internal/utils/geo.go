package utils

import (
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/oschwald/geoip2-golang"
)

type Resolver struct {
	country *geoip2.Reader
	city    *geoip2.Reader
	asn     *geoip2.Reader
}

func NewResolver() *Resolver {

	basePath := os.Getenv("GEOIP_PATH")
	if basePath == "" {
		basePath = "internal/utils/geoDB" // local dev default
	}

	country, err := geoip2.Open(filepath.Join(basePath, "GeoLite2-Country.mmdb"))
	if err != nil {
		log.Printf("failed to open country db: %v", err)
		return nil
	}

	city, err := geoip2.Open(filepath.Join(basePath, "GeoLite2-City.mmdb"))
	if err != nil {
		log.Printf("failed to open city db: %v", err)
		return nil
	}

	asn, err := geoip2.Open(filepath.Join(basePath, "GeoLite2-ASN.mmdb"))
	if err != nil {
		log.Printf("failed to open asn db: %v", err)
		return nil
	}

	return &Resolver{country: country, city: city, asn: asn}
}

func (r *Resolver) CountryCode(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}

	record, err := r.country.Country(ip)
	if err != nil {
		return ""
	}

	return record.Country.IsoCode // "IN", "US", "GB"
}

func (r *Resolver) City(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}

	record, err := r.city.City(ip)
	if err != nil {
		return ""
	}

	return record.City.Names["en"] // "Mumbai", "London" etc
}

func (r *Resolver) ASN(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}

	record, err := r.asn.ASN(ip)
	if err != nil {
		return ""
	}

	return record.AutonomousSystemOrganization // "Google LLC", "Airtel" etc
}

func (r *Resolver) Close() {
	r.country.Close()
	r.city.Close()
	r.asn.Close()
}
