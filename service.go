package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type provider interface {
	name() string
	fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error)
}

func httpGet(ctx context.Context, cli *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ip-json/1.0")
	req.Header.Set("Accept", "application/json")
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 1<<20))
}

// ---- providers ----

type p1 struct{}

func (p1) name() string { return "ip-api.com" }
func (p1) fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error) {
	body, err := httpGet(ctx, cli, fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return nil, err
	}
	var r resp1
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Status == "fail" {
		return nil, fmt.Errorf("ip-api: %s", r.Message)
	}
	out := &JieGuo{
		IP: r.Query, City: r.City, Region: r.RegionName, RegionCode: r.Region,
		Country: r.CountryCode, CountryName: r.Country, Postal: r.Zip,
		Latitude: r.Lat, Longitude: r.Lon, Timezone: r.Timezone,
		ISP: r.ISP, Org: r.Org,
	}
	if r.AS != "" {
		out.ASN = strings.SplitN(r.AS, " ", 2)[0]
	}
	return out, nil
}

type p2 struct{}

func (p2) name() string { return "ipapi.co" }
func (p2) fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error) {
	body, err := httpGet(ctx, cli, fmt.Sprintf("https://ipapi.co/%s/json/", ip))
	if err != nil {
		return nil, err
	}
	var r resp2
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if r.Error {
		return nil, fmt.Errorf("ipapi.co: %s", r.Reason)
	}
	return &JieGuo{
		IP: r.IP, City: r.City, Region: r.Region, RegionCode: r.RegionCode,
		Country: r.Country, CountryName: r.CountryName, Postal: r.Postal,
		Latitude: r.Latitude, Longitude: r.Longitude, Timezone: r.Timezone,
		UTCOffset: r.UTCOffset, Org: r.Org, ASN: r.ASN,
	}, nil
}

type p3 struct{}

func (p3) name() string { return "ipwhois.app" }
func (p3) fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error) {
	body, err := httpGet(ctx, cli, fmt.Sprintf("https://ipwhois.app/json/%s", ip))
	if err != nil {
		return nil, err
	}
	var r resp3
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	if !r.Success {
		return nil, fmt.Errorf("ipwhois: %s", r.Message)
	}
	return &JieGuo{
		IP: r.IP, City: r.City, Region: r.Region, Country: r.CountryCode,
		CountryName: r.Country, Continent: r.Continent,
		Latitude: r.Latitude, Longitude: r.Longitude, Timezone: r.Timezone,
		UTCOffset: r.UTCOffset, ISP: r.ISP, Org: r.Org,
	}, nil
}

type p4 struct{}

func (p4) name() string { return "freeipapi.com" }
func (p4) fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error) {
	body, err := httpGet(ctx, cli, fmt.Sprintf("https://freeipapi.com/api/json/%s", ip))
	if err != nil {
		return nil, err
	}
	var r resp4
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	proxy := r.IsProxy
	return &JieGuo{
		IP: r.IPAddress, City: r.CityName, Region: r.RegionName,
		Country: r.CountryCode, CountryName: r.CountryName, Continent: r.Continent,
		Postal: r.ZipCode, Latitude: r.Latitude, Longitude: r.Longitude,
		Timezone: r.TimeZone, IsProxy: &proxy,
	}, nil
}

type p5 struct{}

func (p5) name() string { return "ip2location.io" }
func (p5) fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error) {
	body, err := httpGet(ctx, cli, fmt.Sprintf("https://api.ip2location.io/?ip=%s", ip))
	if err != nil {
		return nil, err
	}
	var r resp5
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	proxy := r.IsProxy
	out := &JieGuo{
		IP: r.IP, City: r.CityName, Region: r.RegionName,
		Country: r.CountryCode, CountryName: r.CountryName,
		Postal: r.ZipCode, Latitude: r.Latitude, Longitude: r.Longitude,
		Timezone: r.TimeZone, IsProxy: &proxy,
	}
	if r.ASN != "" {
		out.ASN = fmt.Sprintf("AS%s", r.ASN)
	}
	if r.AS != "" {
		out.Org = r.AS
	}
	return out, nil
}

type p6 struct{}

func (p6) name() string { return "db-ip.com" }
func (p6) fetch(ctx context.Context, ip string, cli *http.Client) (*JieGuo, error) {
	body, err := httpGet(ctx, cli, fmt.Sprintf("https://api.db-ip.com/v2/free/%s", ip))
	if err != nil {
		return nil, err
	}
	var r resp6
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	return &JieGuo{
		IP: r.IPAddress, City: r.City, Region: r.StateProv,
		Country: r.CountryCode, CountryName: r.CountryName,
	}, nil
}

// ---- aggregation ----

type lookupResult struct {
	idx  int
	info *JieGuo
}

func lookupAll(ctx context.Context, ip string, providers []provider, cli *http.Client) (*JieGuo, error) {
	ctx, cancel := context.WithTimeout(ctx, cli.Timeout)
	defer cancel()

	ch := make(chan lookupResult, len(providers))
	var wg sync.WaitGroup
	for i, p := range providers {
		wg.Add(1)
		go func(idx int, pv provider) {
			defer wg.Done()
			info, err := pv.fetch(ctx, ip, cli)
			if err != nil {
				slog.Debug("provider error", "name", pv.name(), "ip", ip, "err", err)
				return
			}
			ch <- lookupResult{idx, info}
		}(i, p)
	}
	go func() { wg.Wait(); close(ch) }()

	var results []lookupResult
	for r := range ch {
		results = append(results, r)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("all providers failed")
	}

	return merge(ip, results, len(providers)), nil
}

func merge(ip string, results []lookupResult, total int) *JieGuo {
	out := &JieGuo{IP: ip, Sources: len(results)}

	byIdx := make(map[int]*JieGuo, len(results))
	for _, r := range results {
		byIdx[r.idx] = r.info
	}

	var latSum, lonSum float64
	var coordN int

	for i := 0; i < total; i++ {
		info, ok := byIdx[i]
		if !ok {
			continue
		}
		if out.City == "" {
			out.City = info.City
		}
		if out.Region == "" {
			out.Region = info.Region
		}
		if out.RegionCode == "" {
			out.RegionCode = info.RegionCode
		}
		if out.Country == "" {
			out.Country = info.Country
		}
		if out.CountryName == "" {
			out.CountryName = info.CountryName
		}
		if out.Continent == "" {
			out.Continent = info.Continent
		}
		if out.Postal == "" {
			out.Postal = info.Postal
		}
		if out.Timezone == "" {
			out.Timezone = info.Timezone
		}
		if out.UTCOffset == "" {
			out.UTCOffset = info.UTCOffset
		}
		if out.Org == "" {
			out.Org = info.Org
		}
		if out.ASN == "" {
			out.ASN = info.ASN
		}
		if out.ISP == "" {
			out.ISP = info.ISP
		}
		if out.IsProxy == nil && info.IsProxy != nil {
			out.IsProxy = info.IsProxy
		}
		if info.Latitude != 0 || info.Longitude != 0 {
			latSum += info.Latitude
			lonSum += info.Longitude
			coordN++
		}
	}

	if coordN > 0 {
		out.Latitude = latSum / float64(coordN)
		out.Longitude = lonSum / float64(coordN)
	}

	// reverse dns, best effort
	resolver := &net.Resolver{}
	dnsCtx, dnsCancel := context.WithTimeout(context.Background(), time.Second)
	defer dnsCancel()
	if names, err := resolver.LookupAddr(dnsCtx, ip); err == nil && len(names) > 0 {
		h := names[0]
		if h[len(h)-1] == '.' {
			h = h[:len(h)-1]
		}
		out.Hostname = h
	}

	return out
}

// ---- cache ----

type cache struct {
	mu   sync.RWMutex
	data map[string]cacheItem
	ttl  time.Duration
	max  int
	stop chan struct{}
}

func newCache(ttl time.Duration, max int) *cache {
	c := &cache{
		data: make(map[string]cacheItem),
		ttl:  ttl, max: max,
		stop: make(chan struct{}),
	}
	go c.cleanup()
	return c
}

func (c *cache) get(ip string) (*JieGuo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.data[ip]
	if !ok || time.Now().Unix() > item.expireAt {
		return nil, false
	}
	return item.data, true
}

func (c *cache) set(ip string, val *JieGuo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.data) >= c.max {
		c.evict()
	}
	c.data[ip] = cacheItem{data: val, expireAt: time.Now().Add(c.ttl).Unix()}
}

func (c *cache) setTTL(ip string, val *JieGuo, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.data) >= c.max {
		c.evict()
	}
	c.data[ip] = cacheItem{data: val, expireAt: time.Now().Add(ttl).Unix()}
}

func (c *cache) evict() {
	var oldest string
	var oldestT int64
	first := true
	for k, v := range c.data {
		if first || v.expireAt < oldestT {
			oldest = k
			oldestT = v.expireAt
			first = false
		}
	}
	if !first {
		delete(c.data, oldest)
	}
}

func (c *cache) cleanup() {
	t := time.NewTicker(60 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			c.mu.Lock()
			now := time.Now().Unix()
			for k, v := range c.data {
				if now > v.expireAt {
					delete(c.data, k)
				}
			}
			c.mu.Unlock()
		case <-c.stop:
			return
		}
	}
}

func (c *cache) close() { close(c.stop) }
