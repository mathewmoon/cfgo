package cfgo

import (
  "errors"
  "net/http"
  "time"
  "io/ioutil"
  "fmt"
  "encoding/json"
  "bytes"
)

type Record struct {
  Success bool `json:"success"`
	Errors cferror `json:"errors"`
  Result RecordInfo  `json:"result"`
}

type SingleRecord struct {
  Success  bool `json:"success"`
	Errors  cferror `json:"errors"`
  Result SingleRecordInfo `json:"result"`
}

type SingleRecordInfo struct {
		ID         string    `json:"id"`
		Type       string    `json:"type"`
		Name       string    `json:"name"`
		Content    string    `json:"content"`
		Proxiable  bool      `json:"proxiable"`
		Proxied    bool      `json:"proxied"`
		TTL        int       `json:"ttl"`
		Locked     bool      `json:"locked"`
		ZoneID     string    `json:"zone_id"`
		ZoneName   string    `json:"zone_name"`
		CreatedOn  time.Time `json:"created_on"`
		ModifiedOn time.Time `json:"modified_on"`
		Data       struct {
		} `json:"data"`
}

type RecordInfo []struct {
	Content   string    `json:"content"`
	CreatedOn time.Time `json:"created_on"`
	ID        string    `json:"id"`
	Locked    bool      `json:"locked"`
	Meta      struct {
		AutoAdded           bool `json:"auto_added"`
		ManagedByApps       bool `json:"managed_by_apps"`
		ManagedByArgoTunnel bool `json:"managed_by_argo_tunnel"`
	} `json:"meta"`
	ModifiedOn time.Time `json:"modified_on"`
	Name       string    `json:"name"`
	Proxiable  bool      `json:"proxiable"`
	Proxied    bool      `json:"proxied"`
	TTL        int       `json:"ttl"`
	Type       string    `json:"type"`
	ZoneID     string    `json:"zone_id"`
	ZoneName   string    `json:"zone_name"`
}

type Zone struct {
	Success  bool `json:"success"`
	Errors  cferror `json:"errors"`
  Result ZoneInfo  `json:"result"`
}

type ZoneInfo []struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	DevelopmentMode     int       `json:"development_mode"`
	OriginalNameServers []string  `json:"original_name_servers"`
	OriginalRegistrar   string    `json:"original_registrar"`
	OriginalDnshost     string    `json:"original_dnshost"`
	CreatedOn           time.Time `json:"created_on"`
	ModifiedOn          time.Time `json:"modified_on"`
	Owner               struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		OwnerType string `json:"owner_type"`
	} `json:"owner"`
	Account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"account"`
	Permissions []string `json:"permissions"`
	Plan        struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Price        int    `json:"price"`
		Currency     string `json:"currency"`
		Frequency    string `json:"frequency"`
		LegacyID     string `json:"legacy_id"`
		IsSubscribed bool   `json:"is_subscribed"`
		CanSubscribe bool   `json:"can_subscribe"`
	} `json:"plan"`
	PlanPending struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Price        int    `json:"price"`
		Currency     string `json:"currency"`
		Frequency    string `json:"frequency"`
		LegacyID     string `json:"legacy_id"`
		IsSubscribed bool   `json:"is_subscribed"`
		CanSubscribe bool   `json:"can_subscribe"`
	} `json:"plan_pending"`
	Status      string   `json:"status"`
	Paused      bool     `json:"paused"`
	Type        string   `json:"type"`
	NameServers []string `json:"name_servers"`
}

type User struct {
	Success  bool          `json:"success"`
	Errors   cferror `json:"errors"`
	Result   UserInfo `json:"result"`
}

type UserInfo struct {
		ID                             string    `json:"id"`
		Email                          string    `json:"email"`
		FirstName                      string    `json:"first_name"`
		LastName                       string    `json:"last_name"`
		Username                       string    `json:"username"`
		Telephone                      string    `json:"telephone"`
		Country                        string    `json:"country"`
		Zipcode                        string    `json:"zipcode"`
		CreatedOn                      time.Time `json:"created_on"`
		ModifiedOn                     time.Time `json:"modified_on"`
		TwoFactorAuthenticationEnabled bool      `json:"two_factor_authentication_enabled"`
}

type cferror []struct {
  Code       int `json:"code"`
  ErrorChain []struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
  } `json:"error_chain"`
  Message string `json:"message"`
}

/* The struct for actually making the request */
type Client struct {
  Email string // Account Email address
  Token string // Auth Token
  Domain string // The domain to work with
  lastError cferror //Last Error we encountered
}

/* Gets the last error that was sent from a request to CF */
func (c Client) GetError() string {
  out, err := json.Marshal(c.lastError)
  if err != nil {
      panic (err)
  }
  return "Error: " + string(out)
}

/* Return the details of a DNS zone */
func (c Client) GetZone() (ZoneInfo, error) {
  var z Zone
  var zi ZoneInfo
  endpoint := "https://api.cloudflare.com/client/v4/zones?name=" + c.Domain

  response, err := makeRequest(c, endpoint, "GET", nil)
  if err != nil {
    return zi, err
  }

  json.Unmarshal(response, &z)

  /* If there is an error store the error in the Client object's lastError */
  if !z.Success {
    c.lastError = z.Errors
    return z.Result, errors.New("Request Error")
  }

  return z.Result, nil
}

/* Gets All DNS Records that match `name` and `recordType` */
func (c Client) GetRecord(name string, recordType string) (RecordInfo, error) {
  var r Record
  var ri RecordInfo

  /* Get the Zone ID */
  zone,err := c.GetZone()
  if err != nil {
    return ri, err
  }
  id := zone[0].ID

  /* Make the request for the record */
  endpoint := "https://api.cloudflare.com/client/v4/zones/" + id + "/dns_records?match=all&name=" + name + "&type=" + recordType
  fmt.Print("\n" + endpoint)
  response, err := makeRequest(c, endpoint, "GET", nil)
  if err != nil {
    fmt.Print(err)
    return ri, err
  }

  json.Unmarshal(response, &r)
  fmt.Println(r)
  /* If there is an error store the error in the Client object's lastError */
  if !r.Success {
    fmt.Print(r.Errors)
    c.lastError = r.Errors
    return r.Result, errors.New("Request Error")
  }

  fmt.Print(r)
  return r.Result, nil
}

func (c Client) UpdateRecord(id string, data []byte) (SingleRecordInfo, error) {
  var r SingleRecord
  var ri SingleRecordInfo

  /* Get the Zone ID */
  zone,err := c.GetZone()
  if err != nil {
    return ri, err
  }
  zid := zone[0].ID

  /* Make the request to update the record */
  endpoint := "https://api.cloudflare.com/client/v4/zones/" + zid + "/dns_records/" + id
  response, err := makeRequest(c, endpoint, "PUT", data)
  if err != nil {
    return ri, err
  }

  json.Unmarshal(response, &r)

  /* If there is an error store the error in the Client object's lastError */
  if !r.Success {
    c.lastError = r.Errors
    fmt.Print(r.Errors)
    return ri, errors.New("Request Error")
  }

  fmt.Print(r.Result)
  return r.Result, nil
}

func (c Client) GetUser() UserInfo {
  var u User
  var ui UserInfo

  /* Make the request to update the record */

  endpoint := "https://api.cloudflare.com/client/v4/user"
  response, err := makeRequest(c, endpoint, "GET", data)
  if err != nil {
    return ui, err
  }

  json.Unmarshal(response, &r)

  /* If there is an error store the error in the Client object's lastError */
  if !r.Success {
    c.lastError = r.Errors
    fmt.Print(r.Errors)
    return ui, errors.New("Request Error")
  }

  fmt.Print(r.Result)
  return u.Result, nil

}

/* Make the request */
func makeRequest(c Client, endpoint string, method string, data []byte) ([]byte, error) {
  client := &http.Client{}
  var response []byte
  /* Look up the Zone Identifier From Cloudflare */
  req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(data))
  if err != nil {
    return response, err
  }

  req.Header.Add("X-Auth-Email", c.Email)
  req.Header.Add("X-Auth-Key", c.Token)
  req.Header.Add("Content-Type", "application/json")

  resp, err := client.Do(req)
  if err != nil {
      return response, err
  }

  defer resp.Body.Close()

  responseData, err := ioutil.ReadAll(resp.Body)
  return responseData, err
}
