package main

import (
	"fmt"
	"log"
	"net/http"
  "io"
  "io/ioutil"
  "encoding/json"
  _ "strconv"
  _ "html"
  _ "html/template"
  "os"
  "sync"
  b64 "encoding/base64"
	_ "bytes"

	appengine_log "google.golang.org/appengine/log"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

  "cloud.google.com/go/pubsub"
	_ "google.golang.org/appengine"
  _ "golang.org/x/net/context"
)

var (
	messagesMu sync.Mutex
  countMu sync.Mutex
	count   int
  subscription *pubsub.Subscription
	datasetParentKey string
	datasetNamespace string
)

func main() {
	datasetParentKey = getENV("DATASET_PARENT_KEY")
	datasetNamespace = getENV("DS_NAMESPACE")

	http.HandleFunc("/_ah/health", healthCheckHandler)
  http.HandleFunc("/push", pushHandler)
	//http.HandleFunc("/datastore", datastoreHandler)
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This is main entry for endpoints..")
	})

	log.Print("Starting service.....")
	appengine.Main()
}


type pushRequest struct {
    Message struct {
        Attributes map[string]string
        Data       string `json:"data"`
        message_id     string `json:"message_id"`
        messageId      string `json:"messageId"`
        publish_time   string `json:"publish_time"`
        publishTime    string `json:"publishTime"`
    }
    Subscription string
}

type entityEntryJSONStruct struct {
	Direction string `json:"_direction"`
	Fromst string `json:"_fromst"`
	Last_updt string `json:"_last_updt"`
	Length string `json:"_length"`
	Lif_lat string `json:"_lif_lat"`
	Lit_lat string `json:"_lit_lat"`
	Lit_lon string `json:"_lit_lon"`
	Strheading string `json:"_strheading"`
	Tost string `json:"_tost"`
	Traffic string `json:"_traffic"`
	Segmentid string `json:"segmentid"`
	Start_lon string `json:"start_lon"`
	Street string `json:"street"`
}

type entityEntryDatastoreStruct struct {
	Direction string `datastore:"_direction"`
	Fromst string `datastore:"_fromst"`
	Last_updt string `datastore:"_last_updt"`
	Length string `datastore:"_length"`
	Lif_lat string `datastore:"_lif_lat"`
	Lit_lat string `datastore:"_lit_lat"`
	Lit_lon string `datastore:"_lit_lon"`
	Strheading string `datastore:"_strheading"`
	Tost string `datastore:"_tost"`
	Traffic string `datastore:"_traffic"`
	Segmentid string `datastore:"segmentid"`
	Start_lon string `datastore:"start_lon"`
	Street string `datastore:"street"`
}


func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}

func datastoreHandler(w http.ResponseWriter, r *http.Request, e entityEntryDatastoreStruct) {
	ctx_default := appengine.NewContext(r)

	appengine_log.Infof(ctx_default, "Starting service.....")

	if r.Body == nil {
      appengine_log.Infof(ctx_default, "ERROR: Please send a request body")
      return
  }
  body, err := ioutil.ReadAll(r.Body)
  defer r.Body.Close()
  if err != nil {
    appengine_log.Errorf(ctx_default, "INFO:  Can't read http body ioutil.ReadAll... ")
		return
	}

	// switch context
	ctx, err := appengine.Namespace(ctx_default, "NorthAmerica")
	if err != nil {
		io.WriteString(w, "{\"status\":\"1\", \"message\":\"Can't switch to new Context\"}")
		appengine_log.Errorf(ctx_default, "ERROR: Can't switch to new Context")
		return
	}
	// recover parent key
	parentkey, err := datastore.DecodeKey(datasetParentKey)
	if err != nil {
		appengine_log.Errorf(ctx, "ERROR: Could not decode Parent encoded key:" + string(body))
		io.WriteString(w, "{\"status\":\"1\", \"message\":\"Could not decode Parent encoded key: "+ datasetParentKey +"\"}")
		return
	}

	key := datastore.NewKey(ctx, "entry", "", 0, parentkey)
	fmt.Fprintf(w, "key: " + key.String() + ", key.Namespace: " + key.Namespace() + "\n")

	//log.Print("DEBUG: key: " + key.String())
	//log.Print("DEBUG: entry: " + e.Street)

	go func() {
		newkey, err := datastore.Put(ctx, key, &e)
		if err != nil {
			io.WriteString(w, "{\"status\":\"1\", \"message\":\"Can't execute operation datastore-Put\n\"}")
			appengine_log.Errorf(ctx,"Can't execute operation Datastore PUT: " + err.Error() + "\n")
		}
		appengine_log.Infof(ctx, "DEBUG: saved record Entry type with the key: " + newkey.String() + "\n")
	}()

}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

  if r.Body == nil {
      log.Print("ERROR: Please send a request body")
      return
  }
  body, err := ioutil.ReadAll(r.Body)
  defer r.Body.Close()
  if err != nil {
    log.Printf("INFO:  Can't read http body ioutil.ReadAll... ")
		return
	}
  var msg pushRequest
  if err := json.Unmarshal([]byte(body), &msg); err != nil {
    log.Printf("ERROR: Could not decode body with Unmarshal: %s \n", string(body))
  }
  //log.Printf("DEBUG:  >>>>>  body: %s \n", string(body))
  //log.Printf("DEBUG:  >>>>>  messageId: "    + msg.Message.messageId + "\n")
  sDec, _  := b64.StdEncoding.DecodeString( msg.Message.Data )

  //log.Printf("DEBUG:  >>>>> Message.Data:" + string(sDec) + "\n")
  var data entityEntryJSONStruct
  if err := json.Unmarshal(sDec, &data); err != nil {
    log.Printf("ERROR: Could not decode Message.Data into Entry type with Unmarshal: " + string(sDec) + "\n")
  }

	datastoreHandler(w, r, entityEntryDatastoreStruct(data))

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "{\"status\":\"0\", \"message\":\"ok\"}")

}

// example of use
// c := [100]byte{'a', 'b', 'c'}
// fmt.Println("C: ", len(c), c[:4])
// g := CToGoString(c[:])
// fmt.Println("Go:", len(g), g)
func CToGoString(c []byte) string {
    n := -1
    for i, b := range c {
        if b == 0 {
            break
        }
        n = i
    }
    return string(c[:n+1])
}

func getENV(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("%s environment variable not set.", k)
	}
	return v
}


// eof
