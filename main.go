package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Token struct {
	IP       string
	Account  string
	Password string
}

func readTokens(filename string) ([]Token, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tokens []Token
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 3 {
			tokens = append(tokens, Token{
				IP:       strings.Trim(parts[0], "\""),
				Account:  strings.Trim(parts[1], "\""),
				Password: strings.Trim(parts[2], "\""),
			})
		}
	}
	return tokens, scanner.Err()
}

func callAPI(token Token, apiType, action string, data interface{}, domainValue string) error {
	var method string
	var url string

	switch action {
	case "ADD":
		method = "POST"
		url = fmt.Sprintf("http://%s/nitro/v1/config/%s", token.IP, apiType)
	case "DELETE":
		// 先進行 GET 請求來獲取記錄的詳細資訊
		getURL := fmt.Sprintf("http://%s/nitro/v1/config/%s/%s", token.IP, apiType, domainValue)
		getReq, err := http.NewRequest("GET", getURL, nil)
		if err != nil {
			return err
		}
		getReq.SetBasicAuth(token.Account, token.Password)
		getReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		getResp, err := client.Do(getReq)
		if err != nil {
			return err
		}
		defer getResp.Body.Close()

		getBody, err := ioutil.ReadAll(getResp.Body)
		if err != nil {
			return err
		}

		var getResult map[string]interface{}
		if err := json.Unmarshal(getBody, &getResult); err != nil {
			return err
		}

		// 提取 recordid
		records := getResult[apiType].([]interface{})
		if len(records) == 0 {
			return fmt.Errorf("no records found for domain: %s", domainValue)
		}
		record := records[0].(map[string]interface{})
		recordID := record["recordid"].(string)

		// 構建 DELETE 請求的 URL
		method = "DELETE"
		url = fmt.Sprintf("http://%s/nitro/v1/config/%s/%s?args=recordid:%s", token.IP, apiType, domainValue, recordID)
	case "GET":
		method = "GET"
		url = fmt.Sprintf("http://%s/nitro/v1/config/%s", token.IP, apiType)
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}

	var req *http.Request
	var err error

	if action == "GET" || action == "DELETE" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		jsonData, err := json.Marshal(map[string]interface{}{apiType: data})
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, strings.NewReader(string(jsonData)))
	}

	if err != nil {
		return err
	}

	req.SetBasicAuth(token.Account, token.Password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if action == "GET" {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return err
		}
		log.Printf("Response from %s: %v", token.IP, result)
	} else {
		log.Printf("Response from %s: %s", token.IP, string(body))
	}

	return nil
}

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	rrtype := flag.String("rrtype", "", "Record type (a, aaaa, ns, cname, srv, soa, txt)")
	action := flag.String("action", "GET", "Action to perform (ADD, DELETE, GET)")
	tokenFile := flag.String("f", "nstoken.txt", "Token file")

	// 定義各個記錄類型的參數
	hostname := flag.String("hostname", "", "Hostname for A, AAAA, CNAME records")
	ipaddress := flag.String("ipaddress", "", "IP address for A records")
	ipv6address := flag.String("ipv6address", "", "IPv6 address for AAAA records")
	canonicalname := flag.String("canonicalname", "", "Canonical name for CNAME records")
	domain := flag.String("domain", "", "Domain for MX, NS, SOA, SRV, TXT records")
	mx := flag.String("mx", "", "MX record")
	pref := flag.String("pref", "", "Preference for MX records")
	nameserver := flag.String("nameserver", "", "Nameserver for NS records")
	originserver := flag.String("originserver", "", "Origin server for SOA records")
	contact := flag.String("contact", "", "Contact for SOA records")
	serial := flag.String("serial", "", "Serial for SOA records")
	refresh := flag.Int("refresh", 0, "Refresh for SOA records")
	retry := flag.Int("retry", 0, "Retry for SOA records")
	expire := flag.Int("expire", 0, "Expire for SOA records")
	minimum := flag.Int("minimum", 0, "Minimum for SOA records")
	target := flag.String("target", "", "Target for SRV records")
	priority := flag.String("priority", "", "Priority for SRV records")
	weight := flag.String("weight", "", "Weight for SRV records")
	port := flag.String("port", "", "Port for SRV records")
	var txtStrings stringSlice
	flag.Var(&txtStrings, "txtString", "TXT record string")
	ttl := flag.Int("ttl", 0, "TTL for all records")

	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Println("Usage:")
		fmt.Println("  go run main.go -rrtype <record type> -action <action> [其他參數]")
		fmt.Println("\n參數說明:")
		fmt.Println("  -rrtype：記錄類型（a, aaaa, ns, cname, srv, soa, txt）")
		fmt.Println("  -action：操作類型（ADD, DELETE, GET）")
		fmt.Println("  -f：Token檔案（預設為當前路徑下的 nstoken.txt）")
		fmt.Println("\n各記錄類型所需參數:")
		fmt.Println("\nA 記錄:")
		fmt.Println("  -hostname：主機名")
		fmt.Println("  -ipaddress：IP地址")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype a -action ADD -hostname \"example.com\" -ipaddress \"192.0.2.1\" -ttl 300")
		fmt.Println("\nAAAA 記錄:")
		fmt.Println("  -hostname：主機名")
		fmt.Println("  -ipv6address：IPv6地址")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype aaaa -action ADD -hostname \"example.com\" -ipv6address \"2001:db8::1\" -ttl 300")
		fmt.Println("\nCNAME 記錄:")
		fmt.Println("  -hostname：別名")
		fmt.Println("  -canonicalname：正規名稱")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype cname -action ADD -hostname \"alias.example.com\" -canonicalname \"canonical.example.com\" -ttl 300")
		fmt.Println("\nMX 記錄:")
		fmt.Println("  -domain：域名")
		fmt.Println("  -mx：MX記錄")
		fmt.Println("  -pref：優先級")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype mx -action ADD -domain \"example.com\" -mx \"mail.example.com\" -pref 10 -ttl 300")
		fmt.Println("\nNS 記錄:")
		fmt.Println("  -domain：域名")
		fmt.Println("  -nameserver：名稱伺服器")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype ns -action ADD -domain \"example.com\" -nameserver \"ns1.example.com\" -ttl 300")
		fmt.Println("\nSOA 記錄:")
		fmt.Println("  -domain：域名")
		fmt.Println("  -originserver：原始伺服器")
		fmt.Println("  -contact：聯絡人")
		fmt.Println("  -serial：序列號")
		fmt.Println("  -refresh：刷新時間")
		fmt.Println("  -retry：重試時間")
		fmt.Println("  -expire：過期時間")
		fmt.Println("  -minimum：最小TTL")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype soa -action ADD -domain \"example.com\" -originserver \"ns1.example.com\" -contact \"admin.example.com\" -serial 2023032001 -refresh 3600 -retry 600 -expire 1209600 -minimum 300 -ttl 300")
		fmt.Println("\nSRV 記錄:")
		fmt.Println("  -domain：域名")
		fmt.Println("  -target：目標")
		fmt.Println("  -priority：優先級")
		fmt.Println("  -weight：權重")
		fmt.Println("  -port：端口")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype srv -action ADD -domain \"_service._tcp.example.com\" -target \"target.example.com\" -priority 10 -weight 5 -port 80 -ttl 300")
		fmt.Println("\nTXT 記錄:")
		fmt.Println("  -domain：域名")
		fmt.Println("  -txtString：TXT記錄內容")
		fmt.Println("  -ttl：TTL")
		fmt.Println("  範例：go run main.go -rrtype txt -action ADD -domain \"example.com\" -txtString \"v=spf1 include:_spf.example.com ~all\" -ttl 300")
		os.Exit(1)
	}

	if *rrtype == "" {
		log.Fatalf("Record type is required")
	}

	tokens, err := readTokens(*tokenFile)
	if err != nil {
		log.Fatalf("Error reading tokens: %v", err)
	}

	var apiType string
	var data interface{}
	var domainValue string

	switch *rrtype {
	case "a":
		apiType = "dnsaddrec"
		data = map[string]interface{}{
			"hostname":  *hostname,
			"ipaddress": *ipaddress,
			"ttl":       float64(*ttl),
		}
		domainValue = *hostname
	case "aaaa":
		apiType = "dnsaaaarec"
		data = map[string]interface{}{
			"hostname":    *hostname,
			"ipv6address": *ipv6address,
			"ttl":         float64(*ttl),
		}
		domainValue = *hostname
	case "cname":
		apiType = "dnscnamerec"
		data = map[string]interface{}{
			"aliasname":     *hostname,
			"canonicalname": *canonicalname,
			"ttl":           float64(*ttl),
		}
		domainValue = *hostname
	case "mx":
		apiType = "dnsmxrec"
		data = map[string]interface{}{
			"domain": *domain,
			"mx":     *mx,
			"pref":   *pref,
			"ttl":    float64(*ttl),
		}
		domainValue = *domain
	case "ns":
		apiType = "dnsnsrec"
		data = map[string]interface{}{
			"domain":     *domain,
			"nameserver": *nameserver,
			"ttl":        float64(*ttl),
		}
		domainValue = *domain
	case "soa":
		apiType = "dnssoarec"
		data = map[string]interface{}{
			"domain":       *domain,
			"originserver": *originserver,
			"contact":      *contact,
			"serial":       *serial,
			"refresh":      *refresh,
			"retry":        *retry,
			"expire":       *expire,
			"minimum":      *minimum,
			"ttl":          float64(*ttl),
		}
		domainValue = *domain
	case "srv":
		apiType = "dnssrvrec"
		data = map[string]interface{}{
			"domain":   *domain,
			"target":   *target,
			"priority": *priority,
			"weight":   *weight,
			"port":     *port,
			"ttl":      float64(*ttl),
		}
		domainValue = *domain
	case "txt":
		apiType = "dnstxtrec"
		data = map[string]interface{}{
			"domain": *domain,
			"String": txtStrings,
			"ttl":    float64(*ttl),
		}
		domainValue = *domain
	default:
		log.Fatalf("Unsupported record type: %s", *rrtype)
	}

	for _, token := range tokens {
		err := callAPI(token, apiType, *action, data, domainValue)
		if err != nil {
			log.Printf("Error calling API for %s: %v", token.IP, err)
		}
	}
}
