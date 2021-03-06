package core

import (
	"github.com/gogather/com"
	"io"
	"log"
	"net/http"
)

func HttpStart() {

	http.HandleFunc("/", HandleJsonRpc)

	err := http.ListenAndServe(":1005", nil)
	if err != nil {
		log.Println("ListenAndServe: ", err)
	} else {
		log.Println("Http Server Started!")
	}
}

func HandleJsonRpc(w http.ResponseWriter, r *http.Request) {
	// get request content
	p := make([]byte, r.ContentLength)
	r.Body.Read(p)

	content := string(p)

	log.Println(content)

	json, err := com.JsonDecode(content)

	if err != nil {
		log.Println("not json-rpc format")
		return
	}

	data := json.(map[string]interface{})

	// get system password
	password := C.Get("", "password")

	// parse received password
	passwordRecv, ok := data["password"].(string)
	if !ok {
		result, _ := com.JsonEncode(map[string]interface{}{
			"result": false, //bool, login result
			"msg":    "invalid password, password must be string.",
		})
		io.WriteString(w, result)
		return
	}

	// compare password
	if password != passwordRecv {
		result, _ := com.JsonEncode(map[string]interface{}{
			"result": false, //bool, login failed
		})
		io.WriteString(w, result)
		return
	}

	// trigger controller
	ctrl, exists := RouterMap[data["action"].(string)]
	if !exists {
		log.Println("not exist")
		return
	}
	ctrl.Http(data, w, r)
}
