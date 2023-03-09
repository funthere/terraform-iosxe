package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"html/template"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/ssh"
)

type Item struct {
	ID                  primitive.ObjectID `bson:"_id" json:"id"`
	Host                string             `json:"host"`
	Description         string             `json:"description"`
	Username            string             `json:"username"`
	Password            string             `json:"password"`
	IntfType            string             `json:"type"`
	Number              string             `json:"number"`
	Ipv4Address         string             `json:"ipv4_address"`
	Ipv4AddressMask     string             `json:"ipv4_address_mask"`
	Mtu                 int                `json:"mtu,omitempty"`
	Shutdown            bool               `json:"shutdown,omitempty"`
	ServicePolicyInput  string             `json:"service_policy_input"`
	ServicePolicyOutput string             `json:"service_policy_output"`
}

const templateFile = "api/template/iosxe_interface_ethernet.cfg"
const templateFileDelete = "api/template/iosxe_interface_ethernet_delete.cfg"

// GetItems returns all of the Items that exist in the server
func (s *Service) GetItems(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(map[string]Item{})
	if err != nil {
		log.Println(err)
	}
}

// PostItem handles adding a new Item
func (s *Service) PostItem(w http.ResponseWriter, r *http.Request) {
	defer TimeTrack(time.Now(), "Operations")

	var item Item
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	whiteSpace := regexp.MustCompile(`\s+`)
	if whiteSpace.Match([]byte(item.Host)) {
		http.Error(w, "item names cannot contain whitespace", 400)
		return
	}

	// Load config with template
	commands := loadConfig(item, templateFile)

	// Load SSH config credential
	config := loadSshConfig(item)

	hosts := []string{item.Host}

	// Run the config command
	pushConfig(hosts, commands, config)

	if err != nil {
		log.Printf("error when running command - %s", err)
	}

	item.ID = primitive.NewObjectID()
	_, err = s.db.Collection("config_log").InsertOne(context.TODO(), item)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	log.Printf("added item: %s", item.Host)
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("error sending response - %s", err)
	}

	return
}

// PutItem handles updating an Item with a specific name
func (s *Service) PutItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemName := vars["name"]
	if itemName == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var item Item
	var old Item
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	defer TimeTrack(time.Now(), "Operations")
	// Load config with template
	commands := loadConfig(item, templateFile)

	// Load SSH config credential
	config := loadSshConfig(item)

	hosts := []string{item.Host}

	// Run the config command
	pushConfig(hosts, commands, config)

	if err != nil {
		log.Printf("error when running command - %s", err)
	}

	// Get existing data first, then update.
	filterConfigDB := bson.M{"host": itemName, "intftype": item.IntfType, "number": item.Number}
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	err = s.db.Collection("config_log").FindOne(context.TODO(), filterConfigDB, opts).Decode(&old)

	if err != nil {
		log.Printf(err.Error())
		http.Error(w, fmt.Sprintf("item %v does not exist", itemName), http.StatusBadRequest)
		return
	}

	filterConfigDB = bson.M{"host": itemName, "_id": old.ID}
	item.ID = old.ID
	_, err = s.db.Collection("config_log").ReplaceOne(context.TODO(), filterConfigDB, item)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	log.Printf("updated item: %s", item.Host)
	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Printf("error sending response - %s", err)
	}
}

// DeleteItem handles removing an Item with a specific name
func (s *Service) DeleteItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemName := vars["name"]
	if itemName == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var item Item
	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	defer TimeTrack(time.Now(), "Operations")
	// Load config with template
	commands := loadConfig(item, templateFileDelete)

	// Load SSH config credential
	config := loadSshConfig(item)

	hosts := []string{item.Host}

	// Run the config command
	pushConfig(hosts, commands, config)

	if err != nil {
		log.Printf("error when running command - %s", err)
	}

	// Get existing data first, then update.
	filterConfigDB := bson.M{"host": itemName, "intftype": item.IntfType, "number": item.Number}
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	err = s.db.Collection("config_log").FindOne(context.TODO(), filterConfigDB, opts).Decode(&item)

	if err != nil {
		log.Printf(err.Error())
		http.Error(w, fmt.Sprintf("item %v does not exist", itemName), http.StatusBadRequest)
		return
	}

	filterConfigDB = bson.M{"host": itemName, "_id": item.ID}
	_, err = s.db.Collection("config_log").DeleteOne(context.TODO(), filterConfigDB)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	_, err = fmt.Fprintf(w, "Deleted item with name %s", itemName)
	if err != nil {
		log.Println(err)
	}
}

// GetItem handles retrieving an Item with a specific name
func (s *Service) GetItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemName := vars["name"]
	if itemName == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	var item Item
	filterConfigDB := bson.M{"host": itemName}
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	err := s.db.Collection("config_log").FindOne(context.TODO(), filterConfigDB, opts).
		Decode(&item)
	if err != nil && err.Error() != "mongo: no documents in result" {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if item.Host == "" {
		json.NewEncoder(w).Encode(nil)
		return
	}

	err = json.NewEncoder(w).Encode(item)
	if err != nil {
		log.Println(err)
		return
	}
}

func (s *Service) itemExists(itemName string) bool {
	item := Item{}
	filterConfigDB := bson.M{"host": itemName}
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: -1}})
	err := s.db.Collection("config_log").FindOne(context.TODO(), filterConfigDB, opts).
		Decode(&item)
	if err != nil {
		return false
	}
	if item.Host != "" {
		return true
	}
	return false
}

func loadConfig(item Item, templateFile string) []string {
	commands := []string{}
	t := template.Must(template.ParseFiles(templateFile))

	// 'buf' is an io.Writter to capture the template execution output
	buf := new(bytes.Buffer)
	err := t.Execute(buf, item)
	if err != nil {
		log.Println(err)
		return commands
	}
	commands = strings.Split(buf.String(), "\n")
	commands = removeEmptyStrings(commands)
	// fmt.Println(commands)
	return commands
}

func loadSshConfig(item Item) *ssh.ClientConfig {
	sshConf := ssh.Config{}
	sshConf.SetDefaults()
	sshConf.KeyExchanges = append(
		sshConf.KeyExchanges,
		"diffie-hellman-group-exchange-sha256",
		"diffie-hellman-group-exchange-sha1",
	)

	config := &ssh.ClientConfig{
		User: item.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(item.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Config:          sshConf,
	}
	return config
}

func pushConfig(hosts, commands []string, config *ssh.ClientConfig) error {

	outStrings := make(map[string]string)
	results := make(chan string, 100)

	for _, hostname := range hosts {
		go func(hostname string) {
			results <- executeCmd(hostname, commands, config)
		}(hostname)
	}

	for i := 0; i < len(hosts); i++ {
		res := <-results
		outStrings[hosts[i]] = res
	}

	for _, device_output := range outStrings {
		fmt.Printf("%s", device_output)
		fmt.Printf("\n================================\n\n")
	}
	return nil
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func executeCmd(hostname string, cmds []string, config *ssh.ClientConfig) string {
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	conn, err := ssh.Dial("tcp", hostname, config)
	if err != nil {
		log.Println(err)
	}
	session, err := conn.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	// You can use session.Run() here but that only works
	// if you need a run a single command or you commands
	// are independent of each other.
	err = session.RequestPty("xterm", 80, 40, modes)
	if err != nil {
		log.Fatalf("request for pseudo terminal failed: %s", err)
	}
	stdBuf, err := session.StdoutPipe()
	if err != nil {
		log.Fatalf("request for stdout pipe failed: %s", err)
	}
	stdinBuf, err := session.StdinPipe()
	if err != nil {
		log.Fatalf("request for stdin pipe failed: %s", err)
	}
	err = session.Shell()
	if err != nil {
		log.Fatalf("failed to start shell: %s", err)
	}

	var cmd_output string

	for _, cmd := range cmds {
		errorFlag := false
		stdinBuf.Write([]byte(cmd + "\n"))
		for {
			stdoutBuf := make([]byte, 1000000)
			time.Sleep(time.Millisecond * 40)
			byteCount, err := stdBuf.Read(stdoutBuf)
			if err != nil {
				log.Fatal(err)
			}
			cmd_output += string(stdoutBuf[:byteCount])

			if strings.Contains(string(stdoutBuf[:byteCount]), "syntax error") {
				errorFlag = true
				break
			}
			if !(strings.Contains(string(stdoutBuf[:byteCount]), "More")) {
				break
			}
			stdinBuf.Write([]byte(" "))
		}

		if errorFlag {
			break
		}
	}

	return cmd_output
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
