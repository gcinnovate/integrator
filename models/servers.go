package models

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gcinnovate/integrator/db"
	"github.com/lib/pq"
)

func init() {
	rows, err := db.GetDB().Queryx("SELECT * FROM servers")
	defer rows.Close()

	if err != nil {
		log.Fatalf("Failed to load servers", err)
	}
	ServerMap = make(map[string]Server)
	for rows.Next() {
		srv := &Server{}
		s := &srv.s
		err := rows.StructScan(&s)
		if err != nil {
			log.Fatalln("Server Loading ==>", err)
		}
		// fmt.Printf("=>>>>>>%#v", s)
		ServerMap[strconv.Itoa(int(s.ID))] = *srv

	}
}

// ServerMap is the List of Servers
var ServerMap map[string]Server

// ServerID is the id for the server
type ServerID int64

// Server is our user object
type Server struct {
	s struct {
		ID                      ServerID               `db:"id" 						json:"-"`
		UID                     string                 `db:"uid" 					json:"uid"`
		Name                    string                 `db:"name" 					json:"name"`
		Username                string                 `db:"username"				json:"username"`
		Password                string                 `db:"password"				json:"password"`
		IsProxyServer           bool                   `db:"is_proxy_server"			json:"is_proxy_server"` // whether response is received as is
		SystemType              string                 `db:"system_type"				json:"system_type"`        // the type of system e.g DHIS2, Other is the default
		EndPointType            string                 `db:"endpoint_type"			json:"endpoint_type"`     // e.g /dataValueSets,
		AuthToken               string                 `db:"auth_token"				json:"auth_token"`
		IPAddress               string                 `db:"ipaddress"				json:"ipaddress"` // Usefull for setting Trusted Proxies
		URL                     string                 `db:"url"						json:"url"`
		CCURLS                  pq.StringArray         `db:"cc_urls"					json:"cc_urls"`              // just an additional URL to receive same request
		CallbackURL             string                 `db:"callback_url" 			json:"callback_url"`     // receives response on success call to url
		HTTPMethod              string                 `db:"http_method"				json:"http_method"`       // the HTTP Method used when calling the url
		AuthMethod              string                 `db:"auth_method"				json:"auth_method"`       // the Authentication Method used
		AllowCallbacks          bool                   `db:"allow_callbacks"			json:"allowCallbacks"` // Whether to allow calling sending callbacks
		AllowCopies             bool                   `db:"allow_copies"			json:"allowCopies"`       // Whether to allow copying similar request to CCURLs
		UseAsync                bool                   `db:"use_async" 				json:"use_async"`
		UseSSL                  bool                   `db:"use_ssl" 				json:"use_ssl"`
		ParseResponses          bool                   `db:"parse_responses" 		json:"parseResponses"`
		SSLClientCertKeyFile    string                 `db:"ssl_client_certkey_file"	json:"sslClientCertkeyFile"`
		StartOfSubmissionPeriod string                 `db:"start_submission_period"	json:"startSubmissionPeriod"`
		EndOfSubmissionPeriod   string                 `db:"end_submission_period"	json:"endSubmissionPeriod"`
		XMLResponseXPATH        string                 `db:"xml_response_xpath" 		json:"xml_response_xpath"`
		JSONResponseXPATH       string                 `db:"json_response_xpath" 	json:"json_response_xpath"`
		Suspended               bool                   `db:"suspended" 				json:"suspended"`
		URLParams               map[string]interface{} `db:"url_params" json:"URLParams"`
		Created                 time.Time              `db:"created" 				json:"created"`
		Updated                 time.Time              `db:"updated" 				json:"updated"`
	}
}

// ServerAllowedApps hold servers and servers they allow to communicate with
type ServerAllowedApps struct {
	ID             int64      `db:"id" 				json:"id"`
	ServerID       ServerID   `db:"server_id" 		json:"server_id"`
	AllowedServers []ServerID `db:"allowed_servers" json:"allowed_servers"`
}

// ID return the id of this request
func (s *Server) ID() ServerID { return s.s.ID }

// UID returns the uid of the server/app
func (s *Server) UID() string { return s.s.UID }

// Name ...
func (s *Server) Name() string { return s.s.Name }

// Username ...
func (s *Server) Username() string { return s.s.Username }

// Password ...
func (s *Server) Password() string { return s.s.Password }

// SystemType return the type of system/app it is
func (s *Server) SystemType() string { return s.s.SystemType }

// AuthToken return the Authentication token for this server
func (s *Server) AuthToken() string { return s.s.AuthToken }

// URL returns the URL for the server
func (s *Server) URL() string { return s.s.URL }

// HTTPMethod returns the method used when calling the URL
func (s *Server) HTTPMethod() string { return s.s.HTTPMethod }

// AuthMethod ...
func (s *Server) AuthMethod() string { return s.s.AuthMethod }

// AllowCallbacks returns whether server allows callbacks
func (s *Server) AllowCallbacks() bool { return s.s.AllowCallbacks }

// UseAsync ...
func (s *Server) UseAsync() bool { return s.s.UseAsync }

// CallbackURL return the server callback url
func (s *Server) CallbackURL() string { return s.s.CallbackURL }

// ParseResponses return whether we shold parse the server's responses
func (s *Server) ParseResponses() bool { return s.s.ParseResponses }

// EndOfSubmissionPeriod returns the end of the submission period for the server
func (s *Server) EndOfSubmissionPeriod() string { return s.s.EndOfSubmissionPeriod }

// StartOfSubmissionPeriod returns the start of the submission period for the server
func (s *Server) StartOfSubmissionPeriod() string { return s.s.StartOfSubmissionPeriod }

// Suspended returns whether the server is suspended
func (s *Server) Suspended() bool { return s.s.Suspended }

// CreatedOn return time when Server/App was created
func (s *Server) CreatedOn() time.Time { return s.s.Created }

// UpdatedOn return time when server/app was updated
func (s *Server) UpdatedOn() time.Time { return s.s.Updated }

// GetServerByID returns server object using id
func GetServerByID(id int64) Server {
	srv := Server{}
	err := db.GetDB().Get(&srv.s, "SELECT * FROM servers WHERE id = $1", id)

	if err != nil {
		fmt.Printf("Error geting server: [%v]", err)
		return Server{}
	}
	return srv

}
